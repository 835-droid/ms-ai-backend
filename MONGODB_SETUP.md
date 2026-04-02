# MongoDB Setup Instructions

## Option 1: Use Docker (Recommended)
sudo docker run -d --name msai-mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=changeme \
  -e MONGO_INITDB_DATABASE=MSAIDB \
  mongo:6.0

## Option 2: Use Docker Compose
docker-compose up mongodb -d

## Option 3: Install MongoDB Community Edition
# Follow: https://docs.mongodb.com/manual/tutorial/install-mongodb-on-ubuntu/

## Current Configuration
- MONGO_URI: mongodb://admin:changeme@localhost:27017/MSAIDB?authSource=admin
- DB_TYPE: hybrid (will work with both MongoDB and PostgreSQL)

## Testing
# Test MongoDB connection:
mongosh mongodb://admin:changeme@localhost:27017/MSAIDB

# Test server:
./bin/server

