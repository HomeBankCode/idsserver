import os
import sys
import shutil

if __name__ == "__main__":

    start_dir = sys.argv[1]
    output_dir = sys.argv[2]

    for root, dirs, files in os.walk(start_dir):
        csv_files = [file for file in files if file.endswith(".csv")]
        cha_files = [file for file in files if file.endswith("_idslabel.cha")]

        if len(csv_files) == 1 and len(cha_files) == 1:
            csv_file = csv_files[0]
            cha_file = cha_files[0]

            shutil.copy(os.path.join(root, csv_file), os.path.join(output_dir, csv_file))
            shutil.copy(os.path.join(root, cha_file), os.path.join(output_dir, cha_file))

