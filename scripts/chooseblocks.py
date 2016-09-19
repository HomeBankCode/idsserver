import os
import sys
import csv

import pyclan




if __name__ == "__main__":

    start_dir = sys.argv[1]

    for root, dirs, files in os.walk(start_dir):
        
