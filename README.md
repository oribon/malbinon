# malbinon

sudo certbot certonly --standalone  --register-unsafely-without-email

cp -rL /etc/letsencrypt/live/<HOST_NAME>/* /etc/crts/

sudo docker run -v /etc/crts/:/etc/crts/ -v /var/run/docker.sock:/var/run/docker.sock -p 443:443 -it malbinon:latest
