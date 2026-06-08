#!/bin/bash

echo "Génération des certificats SSL..."

openssl req -x509 -newkey rsa:4096 \
  -keyout key.pem \
  -out cert.pem \
  -days 365 \
  -nodes \
  -subj "//C=FR\ST=France\L=Paris\O=Forum-JS\CN=localhost"

echo "Certificats générés : cert.pem et key.pem"