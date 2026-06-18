build-app:
	docker build -t sanitar/sso:app-v1.0 -f build/app/Dockerfile .
	
build-migrator:
	docker build -t sanitar/sso:migrator-v1.0 -f build/migrator/Dockerfile .
	
build-admin:
	docker build -t sanitar/sso:admin-v1.0 -f build/admin/Dockerfile .
	
push-images:
	docker push sanitar/sso:app-v1.0
	docker push sanitar/sso:migrator-v1.0
	docker push sanitar/sso:admin-v1.0

k8s-run:
	kubectl apply -f deploy/k8s
	
k8s-remove:
	kubectl delete -f deploy/k8s