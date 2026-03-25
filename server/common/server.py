import socket
import logging
import signal
from common.protocol import recv_message, send_confirmation, send_winners
from common.utils import Bet, store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog, clients_total):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._agencies_done = set()
        self._sorteo_done = False
        self._clients_total = clients_total
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, _signum, _frame):
        logging.info("action: shutdown | result: in_progress")
        self._running = False
        try:
            self._server_socket.close()
            logging.info("action: shutdown | result: success")
        except OSError as e:
            logging.error(f"action: shutdown | result: fail | error: {e}")

    def run(self):
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError:
                break

    def __handle_client_connection(self, client_sock):
        agency_id = None
        try:
            while True:
                try:
                    msg = recv_message(client_sock)
                except OSError:
                    logging.warning("action: recv_message | result: client disconnected unexpectedly")
                    break

                if msg is None:
                    if agency_id:
                        self.__handle_eof(agency_id)
                    break

                if msg[0] == "WINNERS":
                    self.__handle_winners_query(client_sock, msg[1])
                    break

                agency_id = msg[0].agency
                self.__handle_batch(client_sock, msg)
        finally:
            client_sock.close()

    def __handle_batch(self, client_sock, bets_info):
        bets = [Bet(b.agency, b.nombre, b.apellido, b.dni, b.nacimiento, b.numero)
                for b in bets_info]
        try:
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            send_confirmation(client_sock, "OK")
        except Exception as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)} | error: {e}')
            send_confirmation(client_sock, "ERROR")

    def __handle_eof(self, agency_id):
        self._agencies_done.add(agency_id)
        if len(self._agencies_done) == self._clients_total:
            logging.info("action: sorteo | result: success")
            self._sorteo_done = True

    def __handle_winners_query(self, client_sock, agency_id):
        if not self._sorteo_done:
            send_confirmation(client_sock, "WAIT")
            return
        winners = [bet.document for bet in load_bets()
                   if str(bet.agency) == str(agency_id) and has_won(bet)]
        send_winners(client_sock, winners)

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
