import socket
import logging
import signal
from common.protocol import recv_batch, send_confirmation
from common.utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
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
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError:
                break


    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            while True:
                try:
                    bets_info = recv_batch(client_sock)
                except OSError:
                    logging.warning("action: recv_batch | result: client disconnected unexpectedly")
                    break

                if bets_info is None:
                    break

                bets = [Bet(b.agency, b.nombre, b.apellido, b.dni, b.nacimiento, b.numero)
                        for b in bets_info]
                try:
                    store_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    send_confirmation(client_sock, "OK")
                except Exception as e:
                    logging.error(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)} | error: {e}')
                    send_confirmation(client_sock, "ERROR")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
