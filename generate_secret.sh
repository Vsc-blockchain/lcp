#! /bin/bash

openssl genpkey -algorithm RSA -out ./enclave/Enclave_private.pem -pkeyopt rsa_keygen_bits:3072 -pkeyopt rsa_keygen_pubexp:3
