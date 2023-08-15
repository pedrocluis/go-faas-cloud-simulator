# Import the modules
import matplotlib.pyplot as plt
import pandas as pd

# Read the file and assign the X
data = pd.read_csv('data.csv')
df = pd.DataFrame(data)

X = list(df.iloc[:, 0])
Y = list(df.iloc[:, 1])

plt.bar(X, Y, color='g', width=1.0)
plt.title("Invocations")
plt.xlabel("Second")
plt.ylabel("Number of invocations")

plt.show()

Y = list(df.iloc[:, 2])

plt.bar(X, Y, color='r', width=1.0)
plt.title("Failed Invocations")
plt.xlabel("Second")
plt.ylabel("Number of failed invocations")

plt.show()

Y = list(df.iloc[:, 3])

plt.bar(X, Y, color='b', width=1.0)
plt.title("Cold Starts")
plt.xlabel("Second")
plt.ylabel("Number of cold invocations")

plt.show()

Y = list(df.iloc[:, 4])

plt.bar(X, Y, color='y', width=1.0)
plt.title("Warm Starts")
plt.xlabel("Second")
plt.ylabel("Number of warm invocations")

plt.show()

Y = list(df.iloc[:, 7])

plt.bar(X, Y, color="orange", width=1.0)
plt.title("Lukewarm starts")
plt.xlabel("Second")
plt.ylabel("Number of lukewarm invocations")

plt.show()

Y = list(df.iloc[:, 5])

plt.bar(X, Y, color="pink", width=1.0)
plt.title("Run memory")
plt.xlabel("Second")
plt.ylabel("Run memory being used")

plt.show()

Y = list(df.iloc[:, 6])

plt.bar(X, Y, color="purple", width=1.0)
plt.title("RAM cache memory")
plt.xlabel("Second")
plt.ylabel("Ram cache memory being used")

plt.show()