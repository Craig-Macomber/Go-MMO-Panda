import struct
import zlib

def gzipXorDelta(data,delta):
    rawDelta=zlib.decompress(delta)
    return [ord(data[i])^ord(rawDelta[i]) for i in xrange(len(data))]

binDeltaMap={
    # keyframe
    0:lambda data,delta:delta,
    # gziped keyframe
    1:lambda data,delta:zlib.decompress(delta),
    2:gzipXorDelta,
    }