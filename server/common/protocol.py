import struct
from common.bet import BetInfo

HEADER_SIZE = 4


def recv_message(sock):
    """Receives a length-prefixed message and returns one of:
    - None: EOF received, client is done sending bets
    - ["WINNERS", agency_id]: winner query for agency_id
    - list[BetInfo]: a batch of bets
    """
    header = recv_all(sock, HEADER_SIZE)
    length = struct.unpack('!I', header)[0]
    payload = recv_all(sock, length).decode('utf-8')

    if payload == "EOF":
        return None

    if payload.startswith("WINNERS|"):
        agency_id = payload.split("|")[1]
        return ["WINNERS", agency_id]

    fields = payload.split('|')
    agency = fields[0]
    bet_fields = fields[1:]

    bets = []
    for i in range(0, len(bet_fields), 5):
        nombre, apellido, dni, nacimiento, numero = bet_fields[i:i+5]
        bets.append(BetInfo(agency, nombre, apellido, dni, nacimiento, numero))
    return bets


def send_confirmation(sock, message: str):
    """Sends a length-prefixed confirmation message. Uses send_all to avoid short-writes."""
    data = message.encode('utf-8')
    header = struct.pack('!I', len(data))
    send_all(sock, header)
    send_all(sock, data)


def send_winners(sock, dnis: list):
    """Sends a length-prefixed list of winner DNIs separated by '|'."""
    payload = "|".join(dnis).encode('utf-8')
    header = struct.pack('!I', len(payload))
    send_all(sock, header)
    send_all(sock, payload)


def send_all(sock, data: bytes):
    """Loops over sock.send until all bytes are sent, avoiding short-writes."""
    sent = 0
    while sent < len(data):
        n = sock.send(data[sent:])
        if n == 0:
            raise OSError("connection closed during send")
        sent += n


def recv_all(sock, length: int) -> bytes:
    """Loops over sock.recv until exactly length bytes are read, avoiding short-reads.
    A partial read is not an error — the remaining bytes will arrive in subsequent calls."""
    buf = b''
    while len(buf) < length:
        chunk = sock.recv(length - len(buf))
        if not chunk:
            raise OSError("connection closed during recv")
        buf += chunk
    return buf
