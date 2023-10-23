# sonic-golang

## Configuração docker sonic

## Rode os comandos abaixo:

### baixar a imagem do sonic:
#### docker pull valeriansaliou/sonic:v1.4.3

### rodar o container docker do sonic:
#### docker run -p 1491:1491 -v /home/paulo/Documentos/Pessoais/sonic-golang/sonic/sonic.cfg:/etc/sonic.cfg -v /home/paulo/Documentos/Pessoais/sonic-golang/sonic/store/:/var/lib/sonic/store/ -d valeriansaliou/sonic:v1.4.3

### Importante configurar de acordo com as pastar do seu computador.

#### configure o sonic na pasta /sonic/sonic.cfg
