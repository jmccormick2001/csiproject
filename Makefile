.PHONY: image
image:
	go build -o csi-driver
	buildah build -t mint:8080/csi-driver:latest -f Dockerfile .
	buildah push --tls-verify=false mint:8080/csi-driver:latest
