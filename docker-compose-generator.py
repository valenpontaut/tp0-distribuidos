import sys


def generate_compose(output_file, num_clients):
    lines = [
        "name: tp0",
        "services:",
        "  server:",
        "    container_name: server",
        "    image: server:latest",
        "    entrypoint: python3 /main.py",
        "    environment:",
        "      - PYTHONUNBUFFERED=1",
        "      - LOGGING_LEVEL=DEBUG",
        "    networks:",
        "      - testing_net",
        "    volumes:",
        "      - ./server/config.ini:/config.ini",
    ]

    for i in range(1, num_clients + 1):
        lines += [
            "",
            f"  client{i}:",
            f"    container_name: client{i}",
            "    image: client:latest",
            "    entrypoint: /client",
            "    environment:",
            f"      - CLI_ID={i}",
            "      - CLI_LOG_LEVEL=DEBUG",
            "    networks:",
            "      - testing_net",
            "    volumes:",
            "      - ./client/config.yaml:/config.yaml",
            "    depends_on:",
            "      - server",
        ]

    lines += [
        "",
        "networks:",
        "  testing_net:",
        "    ipam:",
        "      driver: default",
        "      config:",
        "        - subnet: 172.25.125.0/24",
        "",
    ]

    with open(output_file, "w") as f:
        f.write("\n".join(lines))


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <output_file> <num_clients>")
        sys.exit(1)

    output_file = sys.argv[1]
    num_clients = int(sys.argv[2])
    generate_compose(output_file, num_clients)
    print(f"Generated {output_file} with {num_clients} client(s)")
