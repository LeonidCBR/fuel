#!/bin/bash

# i - include HEADERS to output
# k - insecure (when using self-signed certificate)
# d - data
# X - type of request


curl -i -k \
	-H "Content-Type: application/json" \
	-X POST \
	-d '{"id_vehicle":1,"refueling_date":"2019-10-07","liters":35.1,"cost":1204.158,"odometer":140270}' \
	https://localhost:8585/api/v1/fuel/gasup/create
