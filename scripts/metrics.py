import pandas as pd
import argparse
from sklearn.metrics import accuracy_score, classification_report, precision_recall_fscore_support

parser = argparse.ArgumentParser()
parser.add_argument("inputFolder",help="The folder path")
args = parser.parse_args()
inputFolder=args.inputFolder

with open('/mnt/dataset/labels.csv','r') as f:
    possibleLabels = f.readline().split(',')

data=pd.read_csv(f'{inputFolder}records.csv',header=0,index_col=0)
trueData=pd.read_csv(f'/mnt/dataset/{inputFolder}/dataset.csv',header=0,index_col=0)
labels=data.idxmax(axis="columns",numeric_only=True)
testingCount=0
pred=[]
ground=[]
for filename in labels.index:
    label=labels[filename]
    if data.loc[filename][label]==-1:
        continue
    testingCount+=1
    pred.append(label)
    ground.append(trueData.loc[filename]['label'])
accuracy=accuracy_score(ground, pred)
precision,recall,f1_score,support=precision_recall_fscore_support(ground,pred,labels=possibleLabels,average='macro')
df=pd.DataFrame({'testSampleNum':[testingCount],'accuracy': [accuracy],'precision':[precision],'recall':[recall],'f1_score':[f1_score]})
df.to_csv(path_or_buf='metrics.csv',index=False)