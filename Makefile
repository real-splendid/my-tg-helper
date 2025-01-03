gen-certs:
	openssl req -newkey rsa:2048 -sha256 -nodes -keyout private.key -x509 -days 365 -out public.pem -subj "/C=US/ST=New York/L=Brooklyn/O=Example Brooklyn Company/CN=YOURDOMAIN.EXAMPLE"
	mv private.key cert
	mv public.pem cert

env-from-example:
	cp --no-clobber .env.example .env
