resource "google_sql_database_instance" "vulnerable_example" {
    database_version = "MYSQL_5_7"

    settings {
        tier = "db-f1-micro"

        ip_configuration {
            require_ssl = false  # or unset
        }
    }
 }