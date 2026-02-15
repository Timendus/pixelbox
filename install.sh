#!/bin/sh

sudo cp pixelbox.service /etc/systemd/system/
sudo systemctl enable --now pixelbox.service
sudo systemctl daemon-reload
sudo systemctl restart pixelbox.service
