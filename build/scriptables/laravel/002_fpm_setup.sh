#!/bin/bash
# exit-on-failure=yes
# run_as root

if command -v "php#PHP_VERSION#-fpm" > /dev/null 2>&1; then
    echo "php#PHP_VERSION#-fpm already exists. Nothing to do."
    exit 0
fi

echo "php#PHP_VERSION#-fpm is not installed. Running installation."
SCRIPTABLE::IMPORT php_setup