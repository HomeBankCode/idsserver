import os
import sys
import csv
import random

import pyclan


scrub_intervals = []

def block_passes_criteria(block):
    if (block.num_tier_lines > 10) and \
        (len(block.get_tiers("FAN", "MAN")) >= 2) and\
        (not block_contains_scrub_region(block)):
        return True
    return False

def block_contains_scrub_region(block):
    contains = False
    for interval in scrub_intervals:
        if block.onset <= interval[0] <= block.offset:
            contains = True
        if block.onset <= interval[1] <= block.offset:
            contains = True
        if block.onset >= interval[0] and block.offset <= interval[1]:
            contains = True
    return contains

if __name__ == "__main__":

    start_dir = sys.argv[1]
    output_dir = sys.argv[2]

    for root, dirs, files in os.walk(start_dir):
        cha_files = [file for file in files if file.endswith(".cha")]
        if len(cha_files) == 1:
            cha_file = cha_files[0]
            filepath = os.path.join(root, cha_file)
            csv_path = os.path.join(output_dir, cha_file.replace(".cha", ".csv"))
            new_cha_path = os.path.join(output_dir, cha_file.replace(".cha", "_idslabel.cha"))

            clan_file = pyclan.ClanFile(filepath)

            random_blockrange = clan_file.block_index
            random.shuffle(random_blockrange)

            selected_blocks = []

            scrub_tiers = clan_file.get_tiers("SCR")
            scrub_intervals = []
            if len(scrub_tiers) > 0:
                for interval in scrub_tiers.line_map:
                    scrub_intervals.append([interval.time_onset, interval.time_offset])

            for block_num in random_blockrange:
                block = clan_file.get_conv_block(block_num)

                if block_passes_criteria(block):
                    selected_blocks.append(block.index)
                    if len(selected_blocks) == 60:
                        break

            clan_file.new_file_from_blocks(new_cha_path, selected_blocks[:21])

            with open(csv_path, "wb") as csv_out:
                writer = csv.writer(csv_out)
                writer.writerow(["block_number"])
                selected_blocks = map(str, selected_blocks)
                for number in selected_blocks:
                    writer.writerow([number])
