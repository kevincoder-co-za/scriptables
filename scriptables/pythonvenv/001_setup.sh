#!/bin/bash
# exit-on-failure=yes

sudo apt install python3-venv python3-dev libpq-dev -y
python3 -m venv ~/.venv

~/.venv/bin/pip3 install gunicorn
~/.venv/bin/pip3 install uvicorn
