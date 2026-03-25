import struct
from common.bet import BetInfo

HEADER_SIZE = 4


def recv_bet(sock) -> BetInfo:
    """Receives a length-prefixed bet message. Reads exactly 4 bytes for the
    header, then exactly length bytes for the payload 
    Uses recv_all to avoid short-reads."""
    header = recv_all(sock, HEADER_SIZE)
    length = struct.unpack('!I', header)[0]
    payload = recv_all(sock, length).decode('utf-8')

    agency, nombre, apellido, dni, nacimiento, numero = payload.split('|')
    return BetInfo(agency, nombre, apellido, dni, nacimiento, numero)


def send_confirmation(sock, message: str):
    """Sends a length-prefixed confirmation message. Uses send_all to avoid short-writes."""
    data = message.encode('utf-8')
    header = struct.pack('!I', len(data))
    send_all(sock, header)
    send_all(sock, data)


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
