# Cargar variables de entorno desde .env
include .env
export

# Ejecuta la aplicación en desarrollo
run:
	go run ./cmd/main.go

# Compilar binario
build:
	go build -tags musl -buildvcs=false -o bin/app ./cmd/main.go

clean-mod:
	go clean -cache -modcache
# Ejecutar binario ya compilado
exec:
	./bin/app

# Ejecutar tests unitarios con cobertura y reporte HTML
test:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Ejecutar tests con carrera de gorutinas
race-test:
	go test -race ./...

# Linting con configuración detallada
lint:
	golangci-lint run --color=always --verbose ./...

# Lint modo rápido para pre-commit
lint-fast:
	golangci-lint run --fast --out-format=short ./...

# Test + cobertura + race + linter (modo completo)
check-all: test race-test lint

# Verifica problemas de rendimiento de memoria y gorutinas
profile:
	go test -bench=. -benchmem ./...

# Corre benchmark y muestra top de CPU
cpu-profile:
	go test -cpuprofile cpu.prof -bench=. ./...
	go tool pprof -http=localhost:8080 ./bin/app cpu.prof

# Limpia binarios y archivos de test
clean:
	del /Q bin\* coverage.out coverage.html cpu.prof 2>nul || exit 0


# Corre todo: build, test, lint, profile
all: clean build check-all profile

pre-commit:	clean lint build test

.PHONY: run build exec test race-test lint lint-fast check-all profile cpu-profile clean all pre-commit
