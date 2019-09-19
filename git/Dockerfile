FROM gcr.io/cloud-builders/gcloud

ENV PATH=$PATH:/builder/google-cloud-sdk/bin/

RUN git config --system credential.helper gcloud.sh

RUN wget https://github.com/git-lfs/git-lfs/releases/download/v2.8.0/git-lfs-linux-amd64-v2.8.0.tar.gz \
  && tar -xpsf git-lfs-linux-amd64-v2.8.0.tar.gz \
  && ./install.sh \
  && git lfs install

ENTRYPOINT ["git"]
