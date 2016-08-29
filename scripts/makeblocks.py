import subprocess as sp

import csv
import os
import sys
import random
import re
import datetime
import zipfile
import shutil

from multiprocessing import Process



audio_formats = [".mp3", ".wav"]

block_num_regx = re.compile("(Conversation )+(\d+)")

training = False
reliability = False

class Block:
    def __init__(self, index, clan_file):

        self.index = index
        self.clan_file = clan_file
        self.num_clips = None
        self.clips = []
        self.sliced = False
        self.contains_fan_or_man = False
        self.dont_share = False

class Clip:
    def __init__(self, path, block_index, clip_index):
        self.audio_path = path
        self.parent_audio_path = None
        self.clan_file = None
        self.block_index = block_index
        self.clip_index = clip_index
        self.clip_tier = None
        self.multiline = False
        self.multi_tier_parent = None
        self.start_time = None
        self.offset_time = None
        self.timestamp = None
        self.classification = None
        self.label_date = None
        self.coder = None

class FileGroup:

    def __init__(self, cha, audio):
        self.cha_file = cha
        self.audio_file = audio



class Parser:

    def __init__(self, cha_file, audio_file, clips_dir):

        self.cha_file = cha_file
        self.audio_file = audio_file
        self.clip_blocks = []
        self.clip_directory = clips_dir
        self.audio_file = audio_file
        self.interval_regx = re.compile("\\x15\d+_\d+\\x15")

        self.classification_output = ""
        self.parse_clan(cha_file)

        self.slice_all_man_fan_blocks()

    def parse_clan(self, path):
            conversations = []

            curr_conversation = []
            with open(path, "rU") as file:
                for line in file:
                    if line.startswith("@Bg:\tConversation"):
                        curr_conversation.append(line)
                        continue
                    if curr_conversation:
                        curr_conversation.append(line)
                    if line.startswith("@Eg:\tConversation"):
                        conversations.append(curr_conversation)
                        curr_conversation = []

            conversation_blocks = self.filter_conversations(conversations)


            for index, block in enumerate(conversation_blocks):
                self.clip_blocks.append(self.create_clips(block[1], path, block[0]))

            self.find_multitier_parents()

    def slice_block(self, block):

        clanfilename = block.clan_file.replace(".cha", "")

        class_out_name = block.clan_file.replace(".cha", "_labels.csv")

        all_blocks_path = os.path.join(self.clip_directory, clanfilename)

        self.classification_output = os.path.join(all_blocks_path, str(block.index), class_out_name)

        if not os.path.exists(all_blocks_path):
            os.makedirs(all_blocks_path)

        block_path = os.path.join(all_blocks_path, str(block.index))


        if not os.path.exists(block_path):
            os.makedirs(block_path)

        # showwarning("working directory", "{}".format(os.getcwd()))

        out, err = None, None
        for clip in block.clips:
            command = ["ffmpeg",
                       "-ss",
                       str(clip.start_time),
                       "-t",
                       str(clip.offset_time),
                       "-i",
                       self.audio_file,
                       clip.audio_path,
                       "-y"]

            command_string = " ".join(command)
            print command_string

            pipe = sp.Popen(command, stdout=sp.PIPE, bufsize=10**8)
            out, err = pipe.communicate()

        self.output_classifications(block.index)

    def slice_all_man_fan_blocks(self):
        for block in self.clip_blocks:
            if block.contains_fan_or_man:
                self.slice_block(block)



    def slice_all_randomized_blocks(self):

        for block in self.randomized_blocks:
            self.slice_block(block)

    def create_random_block_range(self):

        self.randomized_blocks = list(self.clip_blocks)
        random.shuffle(self.randomized_blocks)

    def filter_conversations(self, conversations):
        filtered_conversations = []

        last_tier = ""

        block_num = 0
        for conversation in conversations:
            conv_block = []
            for line in conversation:
                if line.startswith("%"):
                    continue
                elif line.startswith("@"):
                    if "Conversation" in line:
                        result = block_num_regx.search(line)
                        if result:
                            block_num = result.group(2)
                    continue
                elif line.startswith("*"):
                    last_tier = line[0:4]
                    conv_block.append(line)
                else:
                    conv_block.append(last_tier+line+"   MULTILINE")
            filtered_conversations.append((block_num, conv_block))
            conv_block = []
            block_num = 0

        return filtered_conversations

    def find_multitier_parents(self):

        for block in self.clip_blocks:
            for clip in block.clips:
                if clip.multiline:
                    self.reverse_parent_lookup(block, clip)

    def reverse_parent_lookup(self, block, multi_clip):
        for clip in reversed(block.clips[0:multi_clip.clip_index-1]):
            if clip.multiline:
                continue
            else:
                multi_clip.multi_tier_parent = clip.timestamp
                return

    def create_clips(self, clips, parent_path, block_index):

        parent_path = os.path.split(parent_path)[1]

        parent_audio_path = os.path.split(self.audio_file)[1]

        block = Block(block_index, parent_path)

        for index, clip in enumerate(clips):

            clip_path = os.path.join(self.clip_directory,
                                     parent_path.replace(".cha", ""),
                                     str(block_index),
                                     str(index+1)+".wav")

            curr_clip = Clip(clip_path, block_index, index+1)
            curr_clip.parent_audio_path = parent_audio_path
            curr_clip.clan_file = parent_path
            curr_clip.clip_tier = clip[1:4]
            if "MULTILINE" in clip:
                curr_clip.multiline = True

            interval_str = ""
            interval_reg_result = self.interval_regx.search(clip)
            if interval_reg_result:
                interval_str = interval_reg_result.group().replace("\x15", "")
                curr_clip.timestamp = interval_str

            time = interval_str.split("_")
            time = [int(time[0]), int(time[1])]

            final_time = self.ms_to_hhmmss(time)

            curr_clip.start_time = str(final_time[0])
            curr_clip.offset_time = str(final_time[2])

            block.clips.append(curr_clip)

        block.num_clips = len(block.clips)

        #self.blocks_to_csv()

        for clip in block.clips:
            if clip.clip_tier == "FAN":
                block.contains_fan_or_man = True
            if clip.clip_tier == "MAN":
                block.contains_fan_or_man = True

        return block

    def ms_to_hhmmss(self, interval):

        x_start = datetime.timedelta(milliseconds=interval[0])
        x_end = datetime.timedelta(milliseconds=interval[1])

        x_diff = datetime.timedelta(milliseconds=interval[1] - interval[0])

        start = ""
        if interval[0] == 0:
            start = "0" + x_start.__str__()[:11] + ".000"
        else:

            start = "0" + x_start.__str__()[:11]
            if start[3] == ":":
                start = start[1:]
        end = "0" + x_end.__str__()[:11]
        if end[3] == ":":
            end = end[1:]

        return [start, end, x_diff]

    def output_classifications(self, block_num):

        #[date, coder, clanfile, audiofile, block, timestamp, clip, tier, label, multi-tier]

        with open(self.classification_output, "wb") as output:
            writer = csv.writer(output)
            writer.writerow(["date", "coder", "clan_file", "audiofile", "block",
                             "timestamp", "clip", "tier", "label", "multi-tier-parent",
                             "dont_share", "training", "reliability"])

            for block in self.clip_blocks:
                if block.index == block_num:
                    dont_share = False
                    if block.dont_share:
                        dont_share = True
                    multitier_parent = None
                    for clip in block.clips:
                        if clip.multiline:
                            multitier_parent = clip.multi_tier_parent
                        else:
                            multitier_parent = "N"

                        writer.writerow([clip.label_date, clip.coder, clip.clan_file,
                                         clip.parent_audio_path, clip.block_index,
                                         clip.timestamp, clip.clip_index,clip.clip_tier,
                                         clip.classification, multitier_parent, dont_share,
                                         training, reliability])
def check_dir(path):
    files = os.listdir(path)

    filtered_files = [file for file in files if not file.startswith(".")]
    if len(filtered_files) == 2\
        and any(".cha" in x for x in files):

        audio = next(file for file in filtered_files if file.endswith(tuple(audio_formats)))
        cha = next(file for file in filtered_files if file.endswith(".cha"))
        abs_audio = os.path.join(path, audio)
        abs_cha = os.path.join(path, cha)
        group = FileGroup(abs_cha, abs_audio)
        return group
    else:
        return None

def zip_block_dirs(clips_dir):
    dirs = [x[0] for x in os.walk(clips_dir)]

    for dir in dirs:
        if dir.count("/") == 2:
            shutil.make_archive(dir, 'zip', dir)


def print_manifest(clips_dir):

    # [clan_file, block_index, path_to_block]
    output = []

    for root, dirs, files in os.walk(clips_dir):
        if any(".zip" in file for file in files):
            for file in files:
                if ".zip" in file:
                    clan_file = os.path.basename(root)
                    block_index = file.replace(".zip", "")
                    block_path = os.path.join(root, file)
                    if training:
                        output.append([clan_file, block_index, block_path, "true", "false"])
                    elif reliability:
                        output.append([clan_file, block_index, block_path, "false", "true"])
                    else:
                        output.append([clan_file, block_index, block_path, "false", "false"])

    with open(os.path.join(clips_dir, "path_manifest.csv"), "wb") as output_file:
        writer = csv.writer(output_file)
        writer.writerow(["clanfile", "block_index", "block_path", "training", "reliability"])
        writer.writerows(output)


def run_parser(cha_file, audio_file, clips_dir):
    parser = Parser(cha_file, audio_file, clips_dir)

if __name__ == "__main__":

    start_dir = sys.argv[1]
    clips_dir = sys.argv[2]

    if "--training" in sys.argv:
        training = True
    if "--reliability" in sys.argv:
        reliability = True


    jobs = []
    for root, dirs, files in os.walk(start_dir):
        if len(dirs) == 0:
            paths = check_dir(root)

            if paths:
                proc = Process(target=run_parser, args=(paths.cha_file, paths.audio_file, clips_dir,))
                proc.start()
                jobs.append(proc)

    for proc in jobs:
        proc.join()

                #parser = Parser(paths.cha_file, paths.audio_file, clips_dir)



    zip_block_dirs(clips_dir)

    print_manifest(clips_dir)
