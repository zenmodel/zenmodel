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

publish: hugo-publish
	@sudo tcb login
	@sudo tcb hosting deploy ./public -e $(TENCENT_ENV_ID)
