import graphlab as gl
from graphlab.toolkits.feature_engineering import DeepFeatureExtractor
from collections import Counter
import os

s3_path = "s3://carwashr/cluster"
image_path = "s3://carwashr/images"
label_path = "s3://carwashr/labels.csv"

ec2 = gl.deploy.ec2_cluster.create(name='carwashr-4',
                                   s3_path=s3_path,
                                   ec2_config=gl.deploy.Ec2Config(),
                                   num_hosts=1)

def train_carwash_network():
    sf = gl.image_analysis.load_images(image_path, "auto", with_path=True,recursive=False, random_order=False)
    sf["id"] = sf["path"].apply(lambda path: os.path.basename(path)[:-4]).astype(int)
    labels = gl.SFrame(label_path)
    sf = sf.join(labels, on="id", how="right")  # only include images with labels
    sf["label"] = sf.apply(lambda row: int(Counter(row["labels"]).most_common()[0][0]))  # use mode label as authoritative label
    print sf["label"]
    train, test = sf.random_split(0.8)
    feature_extractor = gl.feature_engineering.create(sf, DeepFeatureExtractor(feature="image",
                                                                               output_column_name="deep_features"))
    deep_sf = feature_extractor.transform(sf)
    c = gl.classifier.create(deep_sf, target='label', features=["deep_features"])
    return c.evaluate(test)

ob_ec2 = gl.deploy.job.create(train_carwash_network, environment=ec2)
print ob_ec2.get_results()
ec2.stop()
