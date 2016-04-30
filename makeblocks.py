import subprocess as sp

import csv
import os
import sys
import random
import re
import datetime

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



class Parser:

    def __init__(self):

        self.clip_blocks = []

        self.clip_directory = ""

        self.audio_file = ""

        self.interval_regx = re.compile("\\x15\d+_\d+\\x15")

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
                self.clip_blocks.append(self.create_clips(block, path, index+1))

            self.find_multitier_parents()

            # self.block_count_label = Label(self.main_frame,
            #                                text=str(len(conversation_blocks))+\
            #                                " blocks")

            # self.block_count_label.grid(row=27, column=3, columnspan=1)

            # self.create_random_block_range()

    def slice_block(self, block):

        clanfilename = block.clan_file[0:5]

        all_blocks_path = os.path.join(self.clip_directory, clanfilename)

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

    def slice_all_randomized_blocks(self):

        for block in self.randomized_blocks:
            self.slice_block(block)

    def create_random_block_range(self):

        self.randomized_blocks = list(self.clip_blocks)
        random.shuffle(self.randomized_blocks)

    def filter_conversations(self, conversations):
        filtered_conversations = []

        last_tier = ""

        for conversation in conversations:
            conv_block = []
            for line in conversation:
                if line.startswith("%"):
                    continue
                elif line.startswith("@"):
                    continue
                elif line.startswith("*"):
                    last_tier = line[0:4]
                    conv_block.append(line)
                else:
                    conv_block.append(last_tier+line+"   MULTILINE")
            filtered_conversations.append(conv_block)
            conv_block = []

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
                                     parent_path[0:5],
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


if __name__ == "__main__":

    start_dir = sys.argv[1]
    print "hello"

