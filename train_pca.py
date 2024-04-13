from sklearn.decomposition import PCA
from sklearn.decomposition import IncrementalPCA
from sentence_transformers import SentenceTransformer, LoggingHandler, util, evaluation, models, InputExample
from sklearn.preprocessing import normalize
import logging
import os
import gzip
import csv
import random
import numpy as np
import torch
import numpy
import sys
import glob
import re
import concurrent.futures

#New size for the embeddings
NEW_DIM = 192 
NUM_CLUSTERS = 1280
PCA_COMPONENTS_FILE = ("static/pca_%d.npy" % (NEW_DIM))

def train_pca(train_vecs):
    pca = PCA(n_components=NEW_DIM,svd_solver="full")
    pca.fit(train_vecs)
    return pca

def train_ipca(train_vecs, batch_num):
    ipca = IncrementalPCA(n_components=NEW_DIM)
    for i in range(batch_num):
        ipca.partial_fit(train_vecs[i*356317:(i+1)*356317])
        print("complete batch %d/%d" % (i+1, batch_num))
    return ipca

def adjust_precision(vec):
    return numpy.round(numpy.array(vec) * (1<<5))

#train_embeddings = numpy.load("/work/edauterman/clip/deploy.laion.ai/8f83b608504d46bb81708ec86e912220/embeddings/img_emb/img_emb_0.npy")
# train_embeddings = numpy.load("/work/edauterman/private-search/code/embedding/web_msmarco_reduce/web-idx-0.npy")
embed_files_all = glob.glob("static/c4-train-*.npy")
embed_files = []
for i in range(4):
    r = re.compile("static/c4-train-[0-9]+-%d.npy" % i)
    embed_files += list(filter(r.match, embed_files_all))

train_embeddings = [numpy.load(embed_file) for embed_file in embed_files]
print("Loaded npy")
train_embeddings = numpy.concatenate(train_embeddings, axis=0)

train_embeddings = [adjust_precision(embed) for embed in train_embeddings]
print("Loaded and adjusted precision")
pca = train_ipca(train_embeddings, 4)
print("Ran PCA")
numpy.save(PCA_COMPONENTS_FILE, numpy.transpose(pca.components_))
