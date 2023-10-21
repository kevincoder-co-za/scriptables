#!/bin/bash
# exit-on-failure=yes

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Running migrations..."

sudo php#PHP_VERSION# artisan migrate