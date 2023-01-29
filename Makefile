terraform-provider-ssh:
	go mod tidy
	go build -o terraform-provider-ssh

build: terraform-provider-ssh

install: terraform-provider-ssh
	mkdir -p ~/.terraform.d/plugins/patrusoft.at/provider/ssh/0.0.1/linux_amd64
	mv -f terraform-provider-ssh ~/.terraform.d/plugins/patrusoft.at/provider/ssh/0.0.1/linux_amd64 

reinit: 
	rm -f go.mod || true
	rm -f go.sum || true
	go mod init github.com/PaatzDev/terraform-provider-sshtunnel
	go mod tidy