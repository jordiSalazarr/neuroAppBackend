# Makefile

DB_DRIVER = mysql
OUTPUT_DIR = database/dbmodels
PKG_NAME = dbmodels

sqlboiler:
	sqlboiler $(DB_DRIVER) --output $(OUTPUT_DIR) --pkgname $(PKG_NAME) --wipe
