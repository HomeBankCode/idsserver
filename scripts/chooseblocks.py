import os
import sys
import csv
import random

import pyclan

if __name__ == "__main__":

    start_dir = sys.argv[1]

    for root, dirs, files in os.walk(start_dir):
        cha_files = [file for file in files if file.endswith(".cha")]
        if len(cha_files) == 1:
            cha_file = cha_files[0]
            filepath = os.path.join(root, cha_file)
            csv_path = os.path.join(root, cha_file.replace(".cha", ".csv"))
            new_cha_path = os.path.join(root, cha_file.replace(".cha", "_idslabel.cha"))

            clan_file = pyclan.ClanFile(filepath)

            random_blockrange = range(1, clan_file.num_blocks)
            random.shuffle(random_blockrange)

            selected_blocks = []

            for block_num in random_blockrange:
                block = clan_file.get_conv_block(block_num)

                if (block.num_tier_lines > 10) and (len(block.get_tiers("FAN", "MAN")) > 0):
                    selected_blocks.append(block.index)
                    if len(selected_blocks) == 20:
                        break

            clan_file.new_file_from_blocks(new_cha_path, selected_blocks[:21])

            with open(csv_path, "wb") as csv_out:
                writer = csv.writer(csv_out)
                writer.writerow(["block_number"])
                selected_blocks = map(str, selected_blocks)
                for number in selected_blocks:
                    writer.writerow([number])
