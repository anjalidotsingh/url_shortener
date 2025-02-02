# Set database environment variables
export DB_USER=test
export DB_PASSWORD=test
export DB_HOST=localhost
export DB_PORT=3306
export DB_NAME=url_shortner

.PHONY: run_url_shortner
run_url_shortner: setup_db dummy-data
	go run main.go

.PHONY: stop_server
stop_server: 
	docker-compose down
	

.PHONY: setup_db
setup_db:
	docker-compose up -d
	@echo "Waiting for MySQL to be ready..."
	@sleep 20  # Adjust if needed
	@docker exec -i url_shortner_db_container mysql -uroot -pbozcar url_shortner < ./db/migration/01.inital.up.sql
	@echo "Migration completed."
	@docker exec -i url_shortner_db_container mysql -uroot -pbozcar url_shortner -e "SELECT 1 FROM url_mapping LIMIT 1;" || (echo "Table 'url_mapping' does not exist!" && exit 1)
	@docker exec -i url_shortner_db_container mysql -uroot -pbozcar url_shortner -e "SELECT 1 FROM url_count LIMIT 1;" || (echo "Table 'url_count' does not exist!" && exit 1)	

.PHONY: dummy-data
dummy-data:
	@echo "Inserting dummy data into tables..."
	@docker exec -i url_shortner_db_container mysql -uroot -pbozcar url_shortner -e "INSERT INTO url_mapping (actual_url, reference_key) VALUES ('https://www.google.co.in/', 'abc123');"
	@docker exec -i url_shortner_db_container mysql -uroot -pbozcar url_shortner -e "INSERT INTO url_count (domain_name, count) VALUES ('google.co.in', 1);"
	@echo "Dummy data inserted."