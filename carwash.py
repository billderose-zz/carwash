import graphlab as gl
from statistics import mode
import re

loadModel = False
image_path = "./images"


sf = gl.image_analysis.load_images(image_path, "auto", with_path=True, recursive=False, random_order=False)
# sf["image"] = gl.image_analysis.resize(sf["image"], 800, 800, 3)
# sf["id"] = sf["path"].apply(lambda x: remove path, ).astype(int)
sf["id"] = gl.SArray(range(len(sf["image"])))

labels = gl.SFrame("./labels.csv")
sf = sf.join(labels, on="id", how="right")  # only include images with labels
sf["label"] = sf.apply(lambda row: int(mode(row["labels"]))) # use mode label as authoritative label


train, test = sf.random_split(0.8)
if not loadModel:
    nnet = gl.deeplearning.create(train, target='label', features=["image"])
    m = gl.neuralnet_classifier.create(train, target='label', features=["image"], network=nnet,
                                       metric=['accuracy', 'recall@2', 'precision', "fdr", "fp"], validation_set=test, max_iterations=3)
    m.save("./gl.carwash.model")
else:
    m = gl.load_model("./gl.carwash.model")

print m.evaluate(test)
print m.predict(test)
print gl.evaluation.confusion_matrix(test["label"], m.predict)


def similarity(net, im1, im2):
    features = net.extract_features(gl.SFrame({'image': [im1, im2]}))
    return gl.distances.cosine(features[0], features[1])

def print_similarity_matrix(sframe):
    s = ""
    for i, x in enumerate(sframe["image"]):
        if i == 0:
            s += "%d.jpg" % i
        else:
            s += "\t%d.jpg" % i
    print s
    for i, x in enumerate(sframe["image"]):
        s = "%d.jpg" % i
        for y in sf["image"]:
            s += "\t%.2f" % similarity(m, x, y)
        print s
        s = ""
