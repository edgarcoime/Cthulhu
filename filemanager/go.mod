module github.com/edgarcoime/Cthulhu-filemanager

go 1.25.3

replace github.com/edgarcoime/Cthulhu-common => ../common

require (
	github.com/edgarcoime/Cthulhu-common v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require github.com/joho/godotenv v1.5.1 // indirect
