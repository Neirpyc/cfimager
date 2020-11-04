source secrets.txt

BUILD_CONTAINER_ID="2b0bd23a5f1c"

docker network create --driver bridge cfimager-database
docker network create --driver bridge cfimager-compilers-api
docker network create --driver bridge cfimager-compilers
docker network create --driver bridge cfimager-mailer-api
docker network create --driver bridge cfimager-mailer

docker create --name cfimager-mariadb -e MYSQL_ROOT_PASSWORD="$PASSWORD" \
-v /opt/mariadb/data:/var/lib/mysql --network cfimager-database mariadb:10.5

docker create --name cfimager-compilers-spawner -e BUILD_CONTAINER_ID=$BUILD_CONTAINER_ID \
-v /var/run/docker.sock:/var/run/docker.sock --network cfimager-compilers-api \
 --cap-drop ALL \
 --memory=200m --memory-swap=300m --kernel-memory=50m --cpus=2 \
neirpyc/cfimager-compiler-spawner:latest

docker create --name cfimager-mailer -e "ALLOWED_SENDER_DOMAINS=cfimager.neirpyc.ovh" \
  --network cfimager-mailer -p 1587:587 boky/postfix

docker create --name cfimager-mailer-api --network cfimager-mailer-api \
 --cap-drop ALL\
 --memory=40m --memory-swap=10m --kernel-memory=25m --cpus=1 --cpu-shares=512 \
neirpyc/cfimager-mailer:latest

docker create --name cfimager-server -p 127.0.0.1:8042:8070 \
 -e MYSQL_ROOT_PASSWORD="$PASSWORD" -e AUTH_SECRET_SALT="$AUTH_SECRET_SALT" \
 --cap-drop ALL --network cfimager-database\
 --memory=400m --memory-swap=500m --kernel-memory=75m --cpus=3 --cpu-shares=2048 \
 -e HCAPTCHA_SECRET="$HCAPTCHA_SECRET" -e HCAPTCHA_PUBLIC="$HCAPTCHA_PUBLIC" \
neirpyc/cfimager-server:latest

docker network connect cfimager-compilers-api cfimager-server
docker network connect cfimager-compilers cfimager-compilers-spawner
docker network connect cfimager-mailer-api cfimager-server
docker network connect cfimager-mailer cfimager-mailer-api