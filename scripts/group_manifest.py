import csv
import sys

from collections import OrderedDict

if __name__ == "__main__":

    manifest_path = sys.argv[1]

    groups = OrderedDict()

    header = None

    with open(manifest_path, "rU") as input:
        reader = csv.reader(input)
        header = reader.next()

        reliability_rows = []
        training_rows = []

        for row in reader:
            if "reliability" in row[0]:
                reliability_rows.append(row)
                continue
            if "train" in row[0]:
                training_rows.append(row)
                continue
            if row[0] not in groups:
                groups[row[0]] = {"group1": [row+["1"]],
                                  "group2": [],
                                  "group3": []}
                continue
            if len(groups[row[0]]["group1"]) == 10:
                if len(groups[row[0]]["group2"]) == 5:
                    groups[row[0]]["group3"].append(row+["3"])
                else:
                    groups[row[0]]["group2"].append(row+["2"])
            else:
                groups[row[0]]["group1"].append(row+["1"])



    with open("group1_manifest.csv", "wb") as output1:
        writer = csv.writer(output1)
        writer.writerow(header+["group"])

        for row in reliability_rows:
            writer.writerow(row+["1"])
        for row in training_rows:
            writer.writerow(row+["1"])

        for file, groops in groups.iteritems():
            for row in groops["group1"]:
                writer.writerow(row)

    with open("group2_manifest.csv", "wb") as output1:
        writer = csv.writer(output1)
        writer.writerow(header+["group"])

        for row in reliability_rows:
            writer.writerow(row+["2"])
        for row in training_rows:
            writer.writerow(row+["2"])

        for file, groops in groups.iteritems():
            for row in groops["group2"]:
                writer.writerow(row)


    with open("group3_manifest.csv", "wb") as output1:
        writer = csv.writer(output1)
        writer.writerow(header+["group"])

        for row in reliability_rows:
            writer.writerow(row+["3"])
        for row in training_rows:
            writer.writerow(row+["3"])

        for file, groops in groups.iteritems():
            for row in groops["group3"]:
                writer.writerow(row)
