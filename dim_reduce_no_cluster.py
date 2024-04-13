from sklearn.decomposition import PCA
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
from pca import *

#New size for the embeddings
NEW_DIM =192 
PCA_COMPONENTS_FILE = ("static/pca_%d.npy" % NEW_DIM)
PCA_EMBEDDINGS_FILE = ("static/pca_embeddings_%d.npy" % NEW_DIM)

def adjust_precision(vec):
    return numpy.round(numpy.array(vec) * (1<<5))

embed_files_all = glob.glob("static/c4-train-*.npy")
embed_files = []
for i in range(4):
    r = re.compile("static/c4-train-[0-9]+-%d.npy" % i)
    embed_files += list(filter(r.match, embed_files_all))

embeddings = [numpy.load(embed_file) for embed_file in embed_files]
print("Loaded npy")
embeddings = numpy.concatenate(embeddings, axis=0)
# embeddings = numpy.load("static/c4-train-356317-0.npy")
print("npy concatenated")
embeddings = [adjust_precision(embed) for embed in embeddings]
pca_components = numpy.load(PCA_COMPONENTS_FILE)
out_embeddings = numpy.clip(numpy.round(numpy.matmul(embeddings, pca_components)/10), -16, 15)
numpy.save(PCA_EMBEDDINGS_FILE, out_embeddings)
