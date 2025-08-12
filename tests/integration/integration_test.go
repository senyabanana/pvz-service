//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/senyabanana/pvz-service/internal/entity"
	"github.com/senyabanana/pvz-service/internal/repository"
)

const (
	dbUser             = "postgres"
	dbPassword         = "postgres"
	dbName             = "testdb"
	postgresPort       = "5432"
	postgresDriverName = "postgres"
)

func runDBMigration(t *testing.T, migrationURL, dbSource string) {
	t.Helper()

	migration, err := migrate.New(migrationURL, dbSource)
	require.NoError(t, err, "failed to create migration instance")

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err, "failed to apply migrations")
	}

	t.Log("migrations applied successfully")
}

func TestIntegration_PVZFlow(t *testing.T) {
	ctx := context.Background()

	t.Log("Starting PostgreSQL container...")
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.11-alpine3.21",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPassword,
				"POSTGRES_DB":       dbName,
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		},
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(ctx)
	t.Log("PostgreSQL container started")

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, postgresPort)
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, host, port.Port(), dbName)

	t.Log("Connecting to DB...")
	db, err := sqlx.Open(postgresDriverName, dsn)
	require.NoError(t, err, "failed to connect to db")
	defer db.Close()

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	t.Log("Running DB migrations...")
	migrationsPath := "file://../../migrations"
	runDBMigration(t, migrationsPath, dsn)

	pvzRepo := repository.NewPVZPostgres(db)
	receptionRepo := repository.NewReceptionPostgres(db)
	productRepo := repository.NewProductPostgres(db)

	t.Log("Creating PVZ...")
	pvz := &entity.PVZ{
		RegistrationDate: time.Now(),
		City:             entity.CityMoscow,
	}
	err = pvzRepo.CreatePVZ(ctx, pvz)
	require.NoError(t, err)
	t.Logf("PVZ created: %s", pvz.ID)

	t.Log("Creating reception...")
	reception := &entity.Reception{
		DateTime:  time.Now(),
		PVZID:     pvz.ID,
		Status:    entity.StatusInProgress,
		CreatedAt: time.Now(),
	}
	err = receptionRepo.CreateReception(ctx, reception)
	require.NoError(t, err)
	t.Logf("Reception created: %s", reception.ID)

	t.Log("Adding 50 products to reception...")
	for i := 0; i < 50; i++ {
		product := &entity.Product{
			DateTime:    time.Now(),
			Type:        entity.ProductClothing,
			ReceptionID: reception.ID,
		}
		err := productRepo.CreateProduct(ctx, product)
		t.Logf("product type: %s with id: %s added in reception", product.Type, product.ID)
		require.NoError(t, err, fmt.Sprintf("failed to insert product %d", i+1))
	}
	t.Log("50 products added")

	t.Log("Closing reception...")
	err = receptionRepo.CloseReceptionByID(ctx, reception.ID, time.Now())
	require.NoError(t, err)
	t.Log("Reception closed successfully")
}
