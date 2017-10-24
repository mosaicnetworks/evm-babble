nodes = 4

up: conf create start watch

conf: 
	rm -rf conf
	./scripts/build-eth-conf.sh $(nodes)
	./scripts/build-babble-conf.sh $(nodes)

create:
	rm -f ips.dat
	rsync -av --exclude='node_modules' ../web/demo-server/ ../web/demo-server-light
	terraform apply -var-file=secret.tfvars -var "servers=$(nodes)"	

start:
ifdef index
	awk 'FNR==$(index) {system("./scripts/remote-run.sh "$$0" "$$1"")}' ips.dat
else
	awk '{system("./scripts/remote-run.sh "$$0" "$$1"")}' ips.dat  
endif

watch: 
	./scripts/watch.sh

demo: 
	./scripts/demo.sh

stop: 
ifdef index
	awk 'FNR==$(index) {system("./scripts/remote-kill.sh "$$0" "$$1"")}' ips.dat 
else
	awk '{system("./scripts/remote-kill.sh "$$0" "$$1"")}' ips.dat
endif

destroy :
	terraform destroy -var-file=secret.tfvars
	
.PHONY: up conf create start watch demo stop destroy
	