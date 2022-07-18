# go-grafana-tempo

## Requirements/dependencies
- Docker
- Docker-compose
- Golang

## Getting Started

- Starting containers

```sh
make up
```

- Starting API in port `:8080`

```sh
make start
```

- Create traces

```sh
make request
```

- Kill containers

```sh
make down
```

## Exporters

### grafana
- exposed front-end in `http://localhost:3000`