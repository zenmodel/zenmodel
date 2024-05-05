ifneq (,$(wildcard ./.env))
    include .env
    export
endif

depends:
	@hugo mod tidy
	@hugo mod npm pack
	@npm install

hugo-local: depends
	hugo server -w

hugo-publish: depends
	hugo -D

publish-tcb: hugo-publish
	@sudo tcb login
	@sudo tcb hosting deploy ./public -e $(TENCENT_ENV_ID)

publish: hugo-publish
	rsync -avuz --progress ./public/ root@$(TCE_IP):/root/public/
	ssh root@$(TCE_IP) 'sudo nginx -s reload;sudo systemctl restart nginx'