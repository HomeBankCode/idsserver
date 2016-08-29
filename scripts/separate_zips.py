import os
import sys
from shutil import copyfile

def ignore(folders, files):
    ignore_list = []

    for folder in folders:
        ignore_list.append(folder)
    return ignore_list

if __name__ == "__main__":

    start_dir = sys.argv[1]
    end_dir = sys.argv[2]

    if not os.path.exists(end_dir):
        os.makedirs(end_dir)

    # copytree(start_dir, end_dir, ignore=ignore)


    for root, dirs, files in os.walk(start_dir):
        for file in files:
            if ".zip" in file:
                old_path = os.path.join(root, file)
                new_dir = os.path.basename(root)
                new_dir_path = os.path.join(end_dir, new_dir)
                new_path = os.path.join(end_dir, new_dir, file)

                if not os.path.exists(new_dir_path):
                    os.makedirs(new_dir_path)

                # print "old_path: {}".format(old_path)
                # print "new_path: {}".format(new_path)

                copyfile(old_path, new_path)

                #print new_path
