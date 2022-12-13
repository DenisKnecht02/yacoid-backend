docker network create yacoid-backend-network
CALL build_api_image
cd ../ && docker-compose up