# Create

curl -i -X POST localhost:8080/v1/person \
 -H "Content-Type: application/json" \
 -d '{"id":"1","personResource":{"id":"ikkeidag","name":"Alice","alive":true,"age":30}}'

# List

curl localhost:8080/v1/person

# Get one

curl localhost:8080/v1/person/1

# Full update (replace)

curl -i -X PUT localhost:8080/v1/person/1 \
 -H "Content-Type: application/json" \
 -d '{"personResource":{"id":"ikkeidag","name":"Alice","alive":true,"age":32}}'

# Partial update

curl -i -X PATCH localhost:8080/v1/person/1 \
 -H "Content-Type: application/json" \
 -d '{"resource":{"personResource":{"age":32}}}'

# Delete

curl -i -X DELETE localhost:8080/v1/person/1
