FROM fedora
COPY csi-driver /csi-driver
CMD [ "/csi-driver" ]
