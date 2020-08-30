# Malbinon

A web app I wrote to download multiple Docker images simultaneously, in order to move them to a disconnected environment.
Written using Go and JavaScript.
Simplifies the proceess and saves a lot of time:
* Without Malbinon:
  * Connect to a remote server with Docker installed.
  * Pull each image using ```docker pull``` .
  * Save each image using ```docker save```.
* With Malbinon:
  * Enter the images that you want to download.
  * Submit the images and wait for the process to complete, when it finishes you get a directory where you can find your images and download them.

You can also see the images that other people have previously downloaded, saving the time of pulling them again.

## HTTPS simple configuration

sudo certbot certonly --standalone  --register-unsafely-without-email
cp -rL /etc/letsencrypt/live/<HOST_NAME>/* /etc/crts/
docker run -d --restart always -v /etc/crts/:/etc/crts/ -v /var/run/docker.sock:/var/run/docker.sock -v /images:/images -p 443:443 --name malbinon -it malbinon:latest

<img src="https://i.imgur.com/gxsS3lT.jpg" alt="1"/>
<img src="https://i.imgur.com/DDIv82l.jpg" alt="2"/>
<img src="https://i.imgur.com/flwk346.jpg" alt="3"/>
