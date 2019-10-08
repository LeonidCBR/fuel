#!/bin/bash

# i - include HEADERS to output
# k - insecure (when using self-signed certificate)
# d - data
# X - type of request


curl -i -k \
	-H "Content-Type: application/json" \
	-X POST \
	-d '{"dfgl":"gfhj","ttyj":"rrr"}' \
	https://localhost:8585/api/v1/fuel/vehicles/create
