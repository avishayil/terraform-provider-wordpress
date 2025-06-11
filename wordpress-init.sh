#!/bin/sh
# Copyright (c) Avishay Bar
# SPDX-License-Identifier: MIT


set -eu

echo "Starting WordPress initialization..."

# --- 1. Ensure WP-CLI is installed ---
if [ ! -f /usr/local/bin/wp ]; then
  echo "Installing WP-CLI..."
  curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar
  chmod +x wp-cli.phar
  mv wp-cli.phar /usr/local/bin/wp
  echo "WP-CLI installed successfully"
fi

# --- 2. Install MySQL client if needed ---
if ! command -v mysql >/dev/null 2>&1; then
  echo "MySQL client not found - installing..."
  apt-get update \
    && apt-get install -y default-mysql-client \
    && rm -rf /var/lib/apt/lists/*
  echo "MySQL client installed."
fi

# --- 3. Wait for database connection ---
echo "Waiting for database connection..."
max_retries=30
counter=0

while [ "$counter" -lt "$max_retries" ]; do
  if mysqladmin ping -h db -u wordpress -pwordpress >/dev/null 2>&1; then
    echo "Database connection established"
    break
  else
    echo "Database not ready yet, retrying ($counter/$max_retries)"
    sleep 3
    counter=$((counter + 1))
  fi
done

if [ "$counter" -ge "$max_retries" ]; then
  echo "Failed to connect to database after $max_retries attempts"
  exit 1
fi

# --- 4. Configure and install WordPress ---
echo "Installing unzip"
apt-get update && apt-get install -y unzip

echo "Setting up WordPress..."
cd /tmp
curl -O https://wordpress.org/latest.zip
unzip -q latest.zip
rm -rf /var/www/html/*
cp -a wordpress/. /var/www/html/
rm -rf /tmp/wordpress /tmp/latest.zip
echo "WordPress downloaded and extracted to /var/www/html"
cd /var/www/html || exit 1

# Create wp-config.php if it doesn't exist
if [ ! -f wp-config.php ]; then
  echo "Creating wp-config.php"
  wp config create \
    --dbhost=db \
    --dbname=wordpress \
    --dbuser=wordpress \
    --dbpass=wordpress \
    --allow-root \
    --skip-check
fi

# Check if WordPress is installed
if ! wp core is-installed --allow-root >/dev/null 2>&1; then
  echo "Installing WordPress core..."
  wp core install \
    --url=localhost \
    --title="WordPress Test Site" \
    --admin_user=admin \
    --admin_password=admin \
    --admin_email=admin@example.com \
    --allow-root \
    --skip-email

  echo "Updating permalink structure..."
  wp option update permalink_structure "/%postname%/" --allow-root

  echo "Installing initial plugins..."
  wp plugin install hello-dolly --activate --allow-root

  echo "WordPress installation completed."
else
  echo "WordPress is already installed."
fi

echo "WordPress environment ready!"
