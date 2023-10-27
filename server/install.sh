#/usr/bin/bash
echo "######### Instalando asdf #########"
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.13.1
echo "" >> ~/.bashrc
echo '. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc
echo '. "$HOME/.asdf/completions/asdf.bash"' >> ~/.bashrc
echo "" >> ~/.bashrc
. ~/.bashrc
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git

echo "######### Instalando go #########"
asdf install golang latest
asdf global golang latest
go version

echo "######### Compilando #########"
cd go/server/dynamic
go build dynamic.go

echo "######### Rodando #########"
sudo ./dynamic 80
