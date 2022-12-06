FROM jupyter/scipy-notebook:ubuntu-20.04

USER root
RUN apt-get update -y \
    && apt-get install -y netbase \
    && apt-get install -y graphviz \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

USER ${NB_USER}
# mamba installを使いたかったがdatalad pushに失敗するため
# conda installを利用している（2/2時点）
RUN conda install --quiet --yes git==2.35.0 \
    && conda install --quiet --yes git-annex==8.20210903 \
    && conda clean -i -t -y

# install the notebook package etc.
RUN pip install --no-cache --upgrade pip \
    && pip install --no-cache notebook \
    && pip install --no-cache jupyter_contrib_nbextensions \
    && pip install --no-cache git+https://github.com/NII-cloud-operation/Jupyter-LC_run_through \
    && pip install --no-cache git+https://github.com/NII-cloud-operation/Jupyter-multi_outputs \
    && pip install --no-cache datalad==0.17.6 \
    && pip install --no-cache lxml==4.7.1 \
    && pip install --no-cache blockdiag==3.0.0 \
    && pip install --no-cache -U nbformat==5.2.0 \
    && pip install --no-cache black==21.12b0 \
    && pip install --no-cache snakemake \
    && pip install --no-cache boto3 \
    && pip install --no-cache chardet==4.0.0 \
    && pip install --no-cache panel==0.13.1

RUN jupyter contrib nbextension install --user \
    && jupyter nbextensions_configurator enable --user \
    && jupyter run-through quick-setup --user \
    && jupyter nbextension install --py lc_multi_outputs --user \
    && jupyter nbextension enable --py lc_multi_outputs --user


# install Japanese-font (for blockdiag)
ARG font_deb=fonts-ipafont-gothic_00303-18ubuntu1_all.deb
RUN mkdir ${HOME}/.fonts \
    && wget -P ${HOME}/.fonts http://archive.ubuntu.com/ubuntu/pool/universe/f/fonts-ipafont/${font_deb} \
    && dpkg-deb -x ${HOME}/.fonts/${font_deb} ~/.fonts \
    && cp ~/.fonts/usr/share/fonts/opentype/ipafont-gothic/ipag.ttf ~/.fonts/ipag.ttf \
    && rm ${HOME}/.fonts/${font_deb} \
    && rm -rf ${HOME}/.fonts/etc ${HOME}/.fonts/usr \
    && rm .wget-hsts

ARG NB_USER=jovyan
ARG NB_UID=1000

RUN rm -rf ${HOME}/work

# prepare datalad procedure dir
RUN mkdir -p ${HOME}/.config/datalad/procedures

WORKDIR ${HOME}
COPY . ${HOME}

USER root
RUN chown -R ${NB_UID} ${HOME}


##### User Custom Area [Start] ######
# Dockerfileを編集する場合は、「User Custom Area」内に記述してください。
# それ以外のエリアでの記述、また変数名：NB_USERの値は書き換えないでください。ワークフロー機能が正常に働かなくなる可能性があります。
# root権限で実行したい場合は、"USER root"を、一般ユーザ権限で実行したい場合は、 "USER ${NB_USER}"を実行したいコマンド前に記述してください。




###### User Custom Area [End] #######

USER ${NB_USER}
# Specify the default command to run
CMD ["jupyter", "notebook", "--ip", "0.0.0.0"]
