# Up migrations in goose
migrationUp:
	cd sql/schema && goose postgres "postgres://postgres:postgres@localhost:5433/stonk?sslmode=disable" up && cd .. && cd ..

# Down migrations in goose
migrationDown:
	cd sql/schema && goose postgres "postgres://postgres:postgres@localhost:5433/stonk?sslmode=disable" down && cd .. && cd ..

# To connect to stonk db in CLI
databaseDikha: 
	docker exec -it stock-portfolio-tracker-api-postgres-1 psql -U postgres -d stonk