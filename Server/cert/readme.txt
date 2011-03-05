Use these to make the certs:
openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout key.pem -out cert.pem


//openssl rsa -in cert.pem -out public.pem -pubout -outform PEM
//openssl rsa -in cert.pem -out private.pem -outform PEM


Do not used any included certs for anything important! Make your own.