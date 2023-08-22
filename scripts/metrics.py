import pandas as pd
import argparse
from sklearn.metrics import accuracy_score, classification_report

parser = argparse.ArgumentParser()
parser.add_argument("inputFolder",help="The folder path")
args = parser.parse_args()
inputFolder=args.inputFolder

possibleLabels = ['BenignWare', 'Mirai', 'Bashlite', 'Unknown', 'Android', 'Tsunami', 'Dofloo', 'Xorddos', 'Hajime', 'Pnscan']
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
report=classification_report(ground,pred,labels=possibleLabels,output_dict=True)
print(testingCount)
print(accuracy)
print(report)
df=pd.DataFrame({'testSampleNum':[testingCount],'accuracy': [accuracy]})
df.to_csv(path_or_buf='metrics.csv',index=False)