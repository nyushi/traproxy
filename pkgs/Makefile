VERSION=$(shell cat ../VERSION)

DEB_ARCH=$(shell dpkg-architecture -qDEB_BUILD_ARCH)
DPKG=traproxy_$(VERSION)_$(DEB_ARCH).deb

$(DPKG):
	mkdir -p root/usr/sbin
	cp ../traproxy/traproxy ./root/usr/sbin

	fpm -n traproxy -s dir -t deb -v $(VERSION) --after-install deb/after-install --before-remove deb/before-remove --deb-default ./deb/default/traproxy --deb-upstart ./deb/upstart/traproxy -C root usr

clean:
	rm -rf $(DPKG) ./root

.PHONY: dpkg
