import csv
import sys
import os

from pyclan import *



def slice_cha(blocks, file, new_file):

    with open(file, "rU") as input:
        with open(new_file, "wb") as output:
            for index, line in enumerate(input):
                if line.startswith("@") or index < 9:
                    output.write(line)

def read_block_csv(path):
    with open(path, "rU") as input:
        reader = csv.reader(input)
        reader.next()
        return [int(row[0]) for row in reader]


if __name__ == "__main__":

    start_dir = sys.argv[1]

    for root, dirs, files in os.walk(start_dir):
        cha_files = filter(lambda x: x.endswith(".cha"), files)
        if len(cha_files) == 1:
            cha_file = os.path.join(root, cha_files[0])
            new_cha_file = os.path.join(root, cha_file.replace(".cha", "_idslabel.cha"))
            block_csvs = filter(lambda x: x.endswith(".csv"), files)
            if len (block_csvs) == 1:
                block_csv = os.path.join(root, block_csvs[0])
                blocks = read_block_csv(block_csv)

                clan_file = ClanFile(cha_file)
                clan_file.new_file_from_blocks(new_cha_file, blocks)
