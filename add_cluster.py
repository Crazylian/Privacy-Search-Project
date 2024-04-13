from sklearn.cluster import MiniBatchKMeans
import numpy
import pickle
import os
import sys
import glob
import re
import faiss
import concurrent
import threading


NUM_CLUSTERS = 4 * 3500
SCALE_FACTOR = 1000000
MULTI_ASSIGN = 2
DIM=768

def load_f(filename):
    with open(filename, 'r') as f:
        data = numpy.load(filename, allow_pickle=True)
    return data

def test():
    embed_files = ["static/c4-train-356317-2.npy","static/c4-train-356317-3.npy"]
    centroids_file = "static/centroids.npy"
    add_centroids_file= "static/add_centroids.npy"

    data = numpy.empty((0,DIM))
    with concurrent.futures.ThreadPoolExecutor(max_workers=32) as executor:
        future_to_data = [executor.submit(load_f, embed_file) for embed_file in embed_files]
        for i, future in enumerate(concurrent.futures.as_completed(future_to_data)):
            assgin_data = future.result()
            print("Loaded %d" % i)
            if len(data) > 0:
                data = numpy.concatenate((data, assgin_data), axis=0)
            else:
                data = assgin_data

    centroids = numpy.loadtxt(centroids_file)
    kmeans = faiss.Kmeans(DIM,NUM_CLUSTERS,verbose=True)
    kmeans.centroids=centroids.astype(numpy.float32)
    kmeans.train(data.astype(numpy.float32))
    centroids = kmeans.centroids
    print(centroids)
    numpy.savetxt(add_centroids_file, centroids)

    print("Finished kmeans")

if __name__=="__main__":
    test()