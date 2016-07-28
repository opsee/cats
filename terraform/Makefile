ENVIRONMENT ?= production
STACK ?= cats

%-plan: %.tfvars .terraform/terraform.tfstate
ifdef SERVICE
	terraform plan -var-file=$(*).tfvars -target=aws_ecs_service.$(SERVICE) -out $(*).plan
else
	terraform plan -var-file=$(*).tfvars -out $(*).plan
endif

%-apply: %.plan
	terraform apply $(*).plan

plan: $(ENVIRONMENT)-plan

apply: $(ENVIRONMENT)-apply

.terraform:
	mkdir .terraform

.terraform/terraform.tfstate: .terraform
	terraform remote config -backend=s3 -backend-config="bucket=opsee-terraform-state" -backend-config="key=tf/$(STACK)/$(ENVIRONMENT).tfstate" -backend-config="region=us-west-2"

clean:
	rm -f *.plan
	rm -rf .terraform

.PHONY: clean %-apply