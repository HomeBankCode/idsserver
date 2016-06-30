import csv
import sys


def combine_manifests(files):
    data = []
    for file in files:
        with open(file, "rU") as input:
            reader = csv.reader(input)
            reader.next()
            for row in reader:
                data.append(row)

    with open("combined_manifest.csv", "wb") as output:
        writer = csv.writer(output)
        writer.writerow(["clanfile", "block_index", "block_path", "training", "reliability"])
        writer.writerows(data)


if __name__ == "__main__":

    files = sys.argv[1:]
    combine_manifests(files)
