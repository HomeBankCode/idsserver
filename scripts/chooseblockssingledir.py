import os
import sys
import csv
import random

import pyclan


scrub_intervals = []

silence_intervals = {
    "01_06": [
[6967630, 7438634],
[7451760, 8585090],
[9314660, 11976710],
[18733860, 18987490],
[19435080, 19637200],
[19861630, 21495520],
[42861470, 57600140],
    ],

    "15_06": [],

    "21_06": [
[5830842, 11526222],
[30087861, 32147744],
[44985610, 46370173]
    ],

    "32_07": [
[1487310, 1622022],
[6474244, 6945159],
[9143096, 9672713],
[9961175, 11335841],
[11586803, 11900362],
[13488058, 14393973],
[20158889, 24406783],
[26091979, 27097558],
[35278057, 36932387],
[54867811, 54867811]
    ],

    "36_07": [
[11060973, 11317270],
[11679791, 13939478],
[19471482, 26078229],
[34270378, 40984239]
    ],

    "37_06": [
[1000, 8258000],
[37875000, 57599990],
    ],

    "40_07": [
[2440320, 2690710],
[6075590, 10563196],
[11314710, 11541230],
[12672630, 13245810]
    ]
}

def block_passes_criteria(block, file):
    if (block.num_tier_lines > 10) and \
        (len(block.get_tiers("FAN", "MAN")) >= 10) and\
        (not block_contains_scrub_region(block)) and\
        (not block_contains_silence_region(block, file)):
        return True
    return False

def block_contains_scrub_region(block):
    for interval in scrub_intervals:
        if (interval[0] < block.onset) and (interval[1] > block.onset):
            return True
        elif (interval[0] < block.offset) and (interval[1] > block.offset):
            return True
    return False

def block_contains_silence_region(block, file):
    for interval in silence_intervals[file]:
        if (interval[0] < block.onset) and (interval[1] > block.onset):
            return True
        elif (interval[0] < block.offset) and (interval[1] > block.offset):
            return True
    return False

if __name__ == "__main__":

    start_dir = sys.argv[1]
    output_dir = sys.argv[2]

    for root, dirs, files in os.walk(start_dir):
        for file in files:
            print file
            file_key = file[:5]
            cha_file = file

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

                if block_passes_criteria(block, file_key):
                    selected_blocks.append(block.index)
                    if len(selected_blocks) == 25:
                        break

            clan_file.new_file_from_blocks(new_cha_path, selected_blocks[:21])

            with open(csv_path, "wb") as csv_out:
                writer = csv.writer(csv_out)
                writer.writerow(["block_number"])
                selected_blocks = map(str, selected_blocks)
                for number in selected_blocks:
                    writer.writerow([number])
