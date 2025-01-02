gen-certs:
	openssl req \
		-x509 \
		-newkey rsa:2048 \
		-keyout key.pem \
		-days 3560 \
		-subj "/O=Org/CN=Test" \
		-out cert.pem \
		-nodes
	mv key.pem cert
	mv cert.pem cert


env-from-example:
	cp --no-clobber .env.example .env
