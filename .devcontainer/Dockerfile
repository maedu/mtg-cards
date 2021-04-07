# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.140.1/containers/typescript-node/.devcontainer/base.Dockerfile

FROM node

# Install angular-cli && http-server
RUN npm install -g @angular/cli@10 --unsafe-perm \
    && npm install -g http-server    

# Install go-lang
ARG GOLANG_GZ="go1.16.3.linux-arm64.tar.gz"
RUN cd /tmp \ 
    && wget https://golang.org/dl/${GOLANG_GZ}  \
    && tar -C /usr/local -xzf ${GOLANG_GZ}  \
    && echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.bashrc \
    && rm /tmp/${GOLANG_GZ}

# Install go-lang tools for Visual Studio Code
RUN export PATH=$PATH:/usr/local/go/bin \
    && go get -v github.com/acroca/go-symbols \
    && go get -v github.com/cweill/gotests/... \
    && go get -v github.com/davidrjenni/reftools/cmd/fillstruct \
    && go get -v github.com/fatih/gomodifytags \
    && go get -v github.com/godoctor/godoctor \
    && go get -v github.com/haya14busa/goplay/cmd/goplay \
    && go get -v github.com/josharian/impl \
    && go get -v github.com/mdempsky/gocode \
    && go get -v github.com/ramya-rao-a/go-outline \
    && go get -v github.com/rogpeppe/godef \
    && go get -v github.com/sqs/goreturns \
    && go get -v github.com/uudashr/gopkgs/v2/cmd/gopkgs \
    && go get -v golang.org/x/lint/golint \
    && go get -v golang.org/x/tools/cmd/goimports \
    && go get -v golang.org/x/tools/cmd/gorename \
    && go get -v golang.org/x/tools/cmd/guru



# Install heroku cli
#RUN curl https://cli-assets.heroku.com/install.sh | sh

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# [Optional] Uncomment if you want to install an additional version of node using nvm
# ARG EXTRA_NODE_VERSION=10
# RUN su node -c "source /usr/local/share/nvm/nvm.sh && nvm install ${EXTRA_NODE_VERSION}"

# [Optional] Uncomment if you want to install more global node packages
# RUN sudo -u node npm install -g <your-package-list -here>
