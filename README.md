# Malbinon

sudo certbot certonly --standalone  --register-unsafely-without-email

cp -rL /etc/letsencrypt/live/<HOST_NAME>/* /etc/crts/

docker run -d --restart always -v /etc/crts/:/etc/crts/ -v /var/run/docker.sock:/var/run/docker.sock -v /images:/images -p 443:443 --name malbinon -it malbinon:latest

