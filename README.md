# TP0: Docker + Comunicaciones + Concurrencia

**Alumna:** Valentina Llanos Pontaut

**Padrón:** 104413

## Tabla de contenidos

- [Parte 1: Introducción a Docker](#parte-1-introducción-a-docker)
  - [Ejercicio 1](#ejercicio-1)
  - [Ejercicio 2](#ejercicio-2)
  - [Ejercicio 3](#ejercicio-3)
  - [Ejercicio 4](#ejercicio-4)
- [Parte 2: Repaso de Comunicaciones](#parte-2-repaso-de-comunicaciones)
  - [Ejercicio 5](#ejercicio-5)

## Parte 1: Introducción a Docker

### Ejercicio 1

Se creó el script [generar-compose.sh](generar-compose.sh) que recibe dos parámetros y delega la generación al script [docker-compose-generator.py](docker-compose-generator.py):

```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
# Ejemplo:
./generar-compose.sh docker-compose-dev.yaml 5
```

- `<archivo_salida>`: nombre del archivo Docker Compose a generar
- `<cantidad_clientes>`: número de instancias de cliente a incluir

La estructura del Docker Compose generado está inspirada en el ejemplo provisto por la cátedra. A diferencia de ese ejemplo (que define un único cliente), el generador produce N servicios `client1`, `client2`, ..., `clientN` mediante un loop sobre la cantidad de clientes recibida por parámetro, manteniendo el mismo esquema de red y configuración para todos ellos.

### Ejercicio 2

El uso es análogo al explicado en el ejercicio anterior: se ejecuta `generar-compose.sh` con los mismos parámetros para obtener el nuevo archivo de Docker Compose.

La diferencia respecto al ejercicio 1 es que se incorporan volúmenes que montan los archivos de configuración del host dentro de los contenedores:

- El servidor monta `./server/config.ini` en `/config.ini`
- Cada cliente monta `./client/config.yaml` en `/config.yaml`

Esto permite modificar la configuración sin reconstruir las imágenes. Además, se eliminaron las variables de entorno `CLI_LOG_LEVEL` y `LOGGING_LEVEL` que estaban hardcodeadas en el compose, de modo que el nivel de log queda definido exclusivamente por los archivos de configuración (`log.level` en `config.yaml` y `LOGGING_LEVEL` en `config.ini`).

### Ejercicio 3

Se creó el script [validar-echo-server.sh](validar-echo-server.sh) que verifica el funcionamiento del servidor sin instalar herramientas en el host ni exponer puertos. El núcleo del script es el siguiente comando:

```bash
RESPONSE=$(echo "$MESSAGE" | docker run --rm -i --network tp0_testing_net busybox nc server 12345)
```

Cada parte cumple un rol específico:

- `--rm`: elimina el contenedor automáticamente al terminar, sin dejar residuos
- `-i`: mantiene `stdin` abierto para que el contenedor reciba el mensaje via pipe
- `--network tp0_testing_net`: conecta el contenedor a la red interna del compose, lo que le permite resolver el hostname `server` sin necesidad de exponer puertos al host
- `busybox nc server 12345`: usa `netcat` incluido en la imagen `busybox` para conectarse al servidor en el puerto 12345

Si el mensaje recibido coincide con el enviado, el script imprime `action: test_echo_server | result: success`; de lo contrario, `action: test_echo_server | result: fail`.

### Ejercicio 4

#### Cliente (Go)

Se agregó un channel `sigterm` que escucha la señal `SIGTERM` del sistema operativo:

```go
sigterm := make(chan os.Signal, 1)
signal.Notify(sigterm, syscall.SIGTERM)
```

Este channel se consulta en dos puntos del loop mediante un `select`:

1. **Al inicio de cada iteración**, antes de establecer la conexión. El `select` tiene un `case <-sigterm` que interrumpe el loop si llegó la señal, y un `default` que permite continuar si no llegó nada:

```go
select {
case <-sigterm:
    log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
    return
default:
}
```

2. **Durante el sleep entre mensajes**, reemplazando un `time.Sleep` simple por un `select` que puede ser interrumpido:

```go
select {
case <-sigterm:
    log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
    return
case <-time.After(c.config.LoopPeriod):
}
```

Esto garantiza que si llega un SIGTERM mientras el cliente está esperando entre mensajes, la señal se procesa de inmediato en lugar de quedar bloqueada hasta que termine el sleep.

Además, se agregó `defer c.conn.Close()` al momento de crear el socket, asegurando que la conexión se cierre correctamente al salir de la iteración independientemente de cómo termine (éxito, error o señal).

#### Servidor (Python)

Se registró un handler para SIGTERM mediante `signal.signal(signal.SIGTERM, self.__handle_sigterm)` en el constructor. El handler realiza el cierre graceful en tres pasos:

1. Loguea que el shutdown está en progreso
2. Pone `self._running = False` para que el loop principal no acepte más conexiones
3. Cierra el socket del servidor, envuelto en un `try/except OSError` para manejar el caso en que el socket ya estuviera cerrado o en un estado inválido

```python
def __handle_sigterm(self, _signum, _frame):
    logging.info("action: shutdown | result: in_progress")
    self._running = False
    try:
        self._server_socket.close()
        logging.info("action: shutdown | result: success")
    except OSError as e:
        logging.error(f"action: shutdown | result: fail | error: {e}")
```

El loop principal se cambió de `while True` a `while self._running`, de modo que tras recibir SIGTERM no se intentan aceptar nuevas conexiones. El `except OSError` dentro del loop captura la excepción que lanza `accept()` al cerrarse el socket, saliendo limpiamente del loop.

## Parte 2: Repaso de Comunicaciones

### Ejercicio 5

Se modificó la lógica de cliente y servidor para implementar el caso de uso de la Lotería Nacional, donde cada cliente representa una agencia de quiniela que envía una apuesta al servidor central.

#### Protocolo de comunicación

Se implementó un protocolo **length-prefixed** para evitar los fenómenos de short-read y short-write propios de TCP:

```
[ 4 bytes: uint32 big-endian con el largo del payload ][ payload ]
```

El payload serializa los campos de la apuesta separados por `|`:

```
ID|NOMBRE|APELLIDO|DNI|NACIMIENTO|NUMERO
```

Este diseño garantiza que el receptor sepa exactamente cuántos bytes leer sin depender de delimitadores finales ni de que TCP entregue los datos en un único bloque.

#### Separación de responsabilidades

Se crearon módulos dedicados para separar dominio de comunicación:

- `bet.go` / `bet.py`: struct/dataclass `BetInfo` con los campos de la apuesta, sin conocimiento de red
- `protocol.go` / `protocol.py`: funciones de serialización y comunicación (`SendBet`, `RecvConfirmation` en Go; `recv_bet`, `send_confirmation` en Python)

#### Cliente (Go)

Los datos de la apuesta se leen de variables de entorno (`NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO`, `NUMERO`) al iniciar, con validación de tipos:

- `DOCUMENTO` y `NUMERO` deben ser enteros válidos
- `NACIMIENTO` debe tener formato `YYYY-MM-DD`

Para evitar short-writes, `sendAll` loopea sobre `conn.Write` hasta enviar todos los bytes. Para evitar short-reads, `recvAll` loopea sobre `conn.Read` hasta llenar el buffer completo.

Si el servidor no está disponible al momento de conectar, el cliente reintenta hasta 5 veces con 1 segundo de espera entre intentos.

Al recibir la confirmación del servidor se imprime:
```
action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}
```

#### Servidor (Python)

Recibe la apuesta, construye un objeto `Bet` con los campos deserializados y lo persiste mediante `store_bets()`. El mismo mecanismo de `send_all`/`recv_all` se aplica en Python para garantizar envíos y recepciones completas.

Al persistir la apuesta se imprime:
```
action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}
```
