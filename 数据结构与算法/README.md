# 0. 小结

一段时间的数据结构与算法学习大致告一段落了。这里加以总结，梳理。接着继续学习。

数据结构是为算法服务的，算法是为业务服务的。什么样的业务场景，适合使用什么样的算法，这种算法适合使用什么数据结构，这是业务的灵魂。要打实基础，数据结构与算法是绕不过去的。

# 1. 数据结构

分析数据结构离不开增，删，改，查。可以说，业务算法需要数据结构提供的就是合适的增删改查操作。

## 1.1 数组

数组可以说是最简单的数据结构，它顺序存储数据。数组中的数据在栈上分配（如果存储的是指针，则数组分配在堆上）。

### 1.1.1 复杂度分析

数组的增，删，改，查操作的复杂度如下。

**增加**

在数组中增加元素，假设数组空间是足够的。

为保证数组空间的连续性，增加元素是按顺序增加的。那么，先要遍历空闲空间，接着在该空闲空间插入元素。时间复杂度是 O(n)。

**删除**

如果已知元素索引，只需要 O(1) 的复杂度找到要删除的元素，接着删除该元素。

为保证数组空间的连续性，还要对剩余数组进行搬移。时间复杂度是 O(n-k)(k 是元素索引)。

如果元素索引未知，则需要 O(k) 的复杂度先找到元素未知，接着 O(n-k) 的复杂度对元素剩余位置进行搬移。

**改动**

改动也分已知索引和未知索引。已知索引的时间复杂度为 O(1)，未知索引的时间复杂度为 O(n)（其实是查找的时间复杂度）。

**查找**

已知索引的时间复杂度为 O(1)，未知索引的时间复杂度为 O(n)。


时间复杂度分析完了，那么空间复杂度就是 O(n)。操作的一直是连续的预先申请的数组空间。

### 1.1.2 优缺点分析

数组是一段连续的地址空间。所以它的优点也很明显，通过索引查找只需要 O(1) 的时间复杂度，但是由于连续空间，也使得搬移的成本变大，需要 O(n) 的时间复杂度完成插入，删除操作。

由于地址空间连续，CPU 在读取数据时可以批量读取，对性能很有帮助。这是一维数组，在二维数组上，数据是按行连续存储的，如果按列读取元素 CPU 无法批量读取数据，对性能有损耗，这是一个小细节。

还有一点，数组的空间是固定的。那么可能会带来访问越界，以及空间不够的问题。

对于切片类型来说（起底层使用的是数组），如果空间不够，它会申请更大的地址空间，接着将原数组空间拷贝到新数组空间，这一操作相对来说是挺耗时的。尤其是数据量很大的情况。这也间接表明了一个信息，如果数据量大小未知，用数据是不合适的。当然可以使用支持动态扩容的切片来存储数据。

### 1.2.3 练习题

数组虽然易懂，但是在实际做练习题的时候并不容易。

首先，做练习题应该保有的好习惯：
- 思考，先不急于做。思考需求，条件等等，不同条件对应着不同的实现方案。
- 异常 case，有些题的异常 case 是很复杂的，比如 [三树之和](https://leetcode-cn.com/problems/3sum/)。要在思考阶段就要准备异常 case，将异常 case 添加到单元测试中。
- 注释，清晰是特别重要的。复杂度在刷题时很重要，但实际工程上代码写的清晰易读，往往比复杂度来的更为重要。代码是给人看的，这点也要贯彻在刷题的始终。

这是刷题的思路，广义上说也是工程开发的思路，习惯。

当然，只了解数据结构是不够的，数据结构是为算法服务的。什么样的需求适合什么样的“算法”这是要预思考的。

在做题时，整理了一些个人觉得共性的部分。如下：
- 去重，数组去重。即要省去多种组合方式，可以通过排序实现去重的逻辑。这在 [三树之和](https://leetcode-cn.com/problems/3sum/) 上是有体现的。
- 快慢指针，快慢指针可用在环的判断上，如 [环形链表](https://leetcode.cn/problems/linked-list-cycle/)。并且由快慢指针进一步可推倒出“重复节点”，如 [环形链表 II](https://leetcode.cn/problems/linked-list-cycle-ii/description/)。关于快慢指针的变种在数组上也能体现出来，将数组看作链表，判断是否有环，且求出数组的“重复节点”，这是非常考验发散思维的，如 [寻找重复数](https://leetcode.cn/problems/find-the-duplicate-number/description/)。
- 数组的元素和下标，可以看作哈希表。这是纵向扩展，不过使用这种方式需要改动数组。如果需要引入哈希表，却要求用 O(1) 复杂度实现的题目，就要警惕，是否可以将数组当作哈希表了。[两数之和](https://leetcode.cn/problems/two-sum/description/) 是引入哈希表的例子，[缺失的第一个正数](https://leetcode.cn/problems/first-missing-positive/) 是改造数组为哈希表的例子。
- [Boyer-Moore 投票算法] 这是结合数组下标和元素的又一个算法，比较有意思，也挺难想到。题目是 [多数元素](https://leetcode.cn/problems/majority-element/description/)，可以结合着多练练相似题目找找感觉。

## 1.2 链表

链表是不连续的地址空间组成的数据结构，节点用来存储数据，节点间通过指针索引。分为单链表，循环链表，双向链表和双向循环链表。

给出复杂度分析，如下：

| 链表         | 查找        | 删除       | 插入        | 改         |
| ----------- | ----------- | ---------- | ---------- | ---------- |
| 单链表       | O(n)       | O(n)       |  O(n)       | O(n)      |
| 循环链表     | O(n)        |  O(n)     |  O(n)       | O(n)      |
| 双向链表     | O(n)        |  O(n)     |  O(n)       | O(n)      |
| 双向循环链表  | O(n)        |  O(n)     |  O(n)       | O(n)      |

这里的复杂度是包括查找到元素的。比如删除操作，首先是 O(n) 的复杂度查找到要删除的元素，接着以 O(1) 的复杂度删除元素节点。

可以看到，链表存储未知大小元素是非常方便的。相比于数组，它的缺点也很明显，通过指针索引，增加了 CPU 和内存管理器的压力。节点不仅需要存储元素还要存储指针，需要额外的内存空间。

### 1.2.1 练习题

链表的练习题相对来说简洁，且包含指针，是面试常考的类型，把这部分打牢是很有必要的。做了几类练习题发现，双指针使用非常频繁。这里小结如下：
- 删除链表节点：删除链表节点，我们要理清是删除的元素值还是节点。[136. 删除链表的节点](https://leetcode.cn/problems/shan-chu-lian-biao-de-jie-dian-lcof/description/) 是常规的删除链表节点思路，在链表操作时，可使用哨兵节点简化代码。[237. 删除链表中的节点](https://leetcode.cn/problems/delete-node-in-a-linked-list/description/) 是删除链表节点的变种，有意思的是，代码量极少，这题主要考的是思路。
- 双指针法：[83. 删除排序链表中的重复元素](https://leetcode.cn/problems/remove-duplicates-from-sorted-list/description/) 当我们看到排序，并且需要删除重复元素，自然想到双指针法遍历链表。[19. 删除链表的倒数第 N 个结点](https://leetcode.cn/problems/remove-nth-node-from-end-of-list/description/) 但是对于这题，在长度未知情况如何优雅的删除倒数第 N 个结点就不容易了，思路不容易想到。可以利用递归反向获取节点，如解法 [题解](https://leetcode.cn/problems/remove-nth-node-from-end-of-list/submissions/572637275/)，更优雅的是，使用两个间隔为 N 的指针，当快指针走到尾的时候，慢指针刚好指向的是第 N 个结点。这个思路很棒。延展开来，[876. 链表的中间结点](https://leetcode.cn/problems/middle-of-the-linked-list/description/) 也可以使用双指针法找到中间节点。
- 环形链表：涉及到环形链表，可以通过双指针来做。[142. 环形链表 II](https://leetcode.cn/problems/linked-list-cycle-ii/description/) 是一个非常好的环形链表题，这里要知道快慢指针是在同一起跑线开跑的。
- 链表和递归：从尾到头打印链表元素，其实是一个符合先进后出的结构，我们可以将链表拿出来添加到栈中，最后出栈各个元素实现从尾到头打印链表元素。进一步可以想到，递归和栈的执行顺序是一样的，可以使用递归从尾到头打印链表元素，对于小数据量的实现是优雅的。

## 1.3 树

### 1.3.1 二叉查找树

树是一种非线程结构。其中，二叉搜索树，红黑树，堆树，B+ 树是要熟悉的。

理解树要先理解树的三种遍历，前序，中序和后序遍历。对于二叉搜索树来说，其中序遍历既是排序。要注意的树，树的遍历虽然用递归实现，但是遍历树的复杂度为 O(n)，其中 n 表示树中节点的个数。  
树的存储，大部分是用链表存储。对于完全二叉树，如堆树适合在数组中存储。

二叉搜索树，故名思义，主要是为了查找。查找有散列表，可以以 O(1) 的复杂度查找元素，非常高效。但是散列表的查找也有“缺点”，主要有以下几点：
- 散列表是无序的，二叉搜索树不仅适合查找，而且是有序的。
- 散列表底层用数组存储，在内存不足时需要动态扩容。同时，散列表提供的装载因子，会造成一定程度的空间浪费。而且，散列表在空间不足时容易造成散列冲突，在小规模有限数据的查找使用散列表是比较合适的。
- 二叉搜索树具有链表的优点，存储数据无上限。且对查找，排序都有很好的支持。


```
我们在散列表那节中讲过，散列表的插入、删除、查找操作的时间复杂度可以做到常量级的 O(1)，非常高效。而二叉查找树在比较平衡的情况下，插入、删除、查找操作时间复杂度才是 O(logn)，相对散列表，好像并没有什么优势，那我们为什么还要用二叉查找树呢？

我认为有下面几个原因：

第一，散列表中的数据是无序存储的，如果要输出有序的数据，需要先进行排序。而对于二叉查找树来说，我们只需要中序遍历，就可以在 O(n) 的时间复杂度内，输出有序的数据序列。

第二，散列表扩容耗时很多，而且当遇到散列冲突时，性能不稳定，尽管二叉查找树的性能不稳定，但是在工程中，我们最常用的平衡二叉查找树的性能非常稳定，时间复杂度稳定在 O(logn)。

第三，笼统地来说，尽管散列表的查找等操作的时间复杂度是常量级的，但因为哈希冲突的存在，这个常量不一定比 logn 小，所以实际的查找速度可能不一定比 O(logn) 快。加上哈希函数的耗时，也不一定就比平衡二叉查找树的效率高。

第四，散列表的构造比二叉查找树要复杂，需要考虑的东西很多。比如散列函数的设计、冲突解决办法、扩容、缩容等。平衡二叉查找树只需要考虑平衡性这一个问题，而且这个问题的解决方案比较成熟、固定。

最后，为了避免过多的散列冲突，散列表装载因子不能太大，特别是基于开放寻址法解决冲突的散列表，不然会浪费一定的存储空间。

综合这几点，平衡二叉查找树在某些方面还是优于散列表的，所以，这两者的存在并不冲突。我们在实际的开发过程中，需要结合具体的需求来选择使用哪一个。
```

虽然二叉搜索树的插入，查找，删除操作的复杂度为 O(log^n)（和树高相关），不过二叉搜索树的缺点，也是有的。频繁的插入，删除会导致二叉树蜕化，严重的二叉树退化成链表，导致插入，查找的复杂度上升到 O(n)。

要维持二叉搜索树这一平衡结构就需要红黑树，红黑树通过节点的自旋使得树始终保持二叉搜索树在完全二叉树的结构上，不会蜕化。不过红黑树的缺点也是有的，实现复杂，相比于其他树结构，红黑树的插入和删除操作较为复杂，涉及到大量的旋转和平衡调整操作。因此，实际编写代码时容易出错，且维护起来较为困难。


### 1.3.2 练习

树的练习题很多，大部分需要用到遍历，可以通过递归实现。

- 遍历树：根据树遍历的特点，当给定前序和中序或者中序和后序时可以构造一棵树。[106. 从中序与后序遍历序列构造二叉树](https://leetcode.cn/problems/construct-binary-tree-from-inorder-and-postorder-traversal/) 是已知中序和后序确定二叉树的例子。已知前序和中序相对而言更简单，思路也类似，这里就不贴了。
- 二叉树转化：[114. 二叉树展开为链表](https://leetcode.cn/problems/flatten-binary-tree-to-linked-list/description/) 是将二叉树转换为单链表的问题，这是结合遍历和指针的一道题，这题有点意思，我们使用的是新的遍历方式实现的链表转换。[二叉树转换为双链表](./树/tree/convert.go) 这题有点类似，不同的是我们想着递归去做，发现并不好做，转而使用中序遍历结合指针来做这道题，很巧妙。
- 对称二叉树：[101. 对称二叉树](https://leetcode.cn/problems/symmetric-tree/description/) 这里对称，要抓住对称的逻辑。并且，这里有意思的点在于，这里每次递归比较左右子树，判断是否相同来确定是否对称。
- 翻转二叉树：[226. 翻转二叉树](https://leetcode.cn/problems/invert-binary-tree/description/) 要注意，翻转的是什么，这里理清了翻转的是节点，在运用递归去做就不难了。
- 验证二叉树：[98. 验证二叉搜索树](https://leetcode.cn/problems/validate-binary-search-tree/description/) 这题刚开始，我们想用递归去做，后来发现，递归还要确定和其它子树的关系，比较复杂，于是放弃。转而利用二叉搜索树的性质，中序遍历是有序的，我们可以借着这一点用递归结合二叉树的性质去做。可以维护一个指针指向前一个节点，通过比较当前节点和前一个节点确定是否是二叉搜索树。这里我们比较暴力，直接把所有节点加到数组中比较，空间复杂度为 O(n)。
- 子结构判断：[LCR 143. 子结构判断](https://leetcode.cn/problems/shu-de-zi-jie-gou-lcof/description/) 这道题有意思的点在于递归的顺序，当有一个子树判断成功即返回，不需要判断其它子树。这是困扰我们的点，可以通过维护一个 flag 的方式判断是否退出递归。
- 公共祖先：[235. 二叉搜索树的最近公共祖先](https://leetcode.cn/problems/lowest-common-ancestor-of-a-binary-search-tree/description/) 这题首先把二叉搜索树的性质用上，通过几种情况，判断递归路径来确定公共祖先。更进阶的题是 [树的公众节点]，如果节点中有指向父节点的指针，那么可以将问题转换为 [160. 相交链表](https://leetcode.cn/problems/intersection-of-two-linked-lists/description/) 确定公共祖先。如果没有指向父节点的指针，那么可以运用散列表，构建辅助空间，在运用相交链表的求法确定公共祖先。
- 二叉树的深度：[104. 二叉树的最大深度](https://leetcode.cn/problems/maximum-depth-of-binary-tree/description/) 这是另一种形式的递归，理清逻辑应该不难的。
- 路径总和：[112. 路径总和](https://leetcode.cn/problems/path-sum/) 这题可以看作图的广度优先搜索的退化，要维护路径的状态信息需要数据结构，这里结合队列做这道题，应该不难的。这题和二叉树的层序遍历有点类似，都是利用队列做辅助，本质上还是广度优先搜索。


要灵活运用递归还是要思路清晰，如果逻辑太复杂，得考虑换其它思路实现，毕竟递归不是万能得。

## 1.4 栈和队列

栈和队列我们可以说是相当了解的数据结构。但是用栈和队列来做题，则总有一种懵逼的感觉。究其原因，还是题目做少了，不会灵活运用。

- [20. 有效的括号](https://leetcode.cn/problems/valid-parentheses/) 有效的括号要注意，何时入栈，何时出栈，以及为什么想到用栈这种数据结构。这是我们懵逼的核心点。
- [641. 设计循环双端队列](https://leetcode.cn/problems/design-circular-deque/) 这个没什么好讲的，用链表加双端指针实现。

## 1.5 堆树

堆这个数据结构笔试不考，要考这个就太变态了。要注意的是，堆树是一种动态数据结构，要是动态的数据，用堆插入，删除，查找很方便。

要了解建堆的过程，插入，删除堆是怎么做的，复杂度如何。能手写代码出来，并且知道 Top K，中位数和优先级队列的实现原理。

## 1.6 图



# 2. 算法

## 2.1 排序

排序是非常基础的算法。对于冒泡，插入，选择这三种排序的时间复杂度都很高，但较为简单，在小数量排序时可以使用。面试个人感觉除非面试官要求，不然很少会写这三类排序的，因为有复杂度更好的归并排序和快速排序可以用，要把主要精力投在时间复杂度为 O(nlog^n) 的归并和快速排序上。


| 排序         |  时间复杂度    | 原地       | 稳定     |
| -----------  | -----------   | ---------- | ------- |
| 冒泡         |  O(n^2)        | 是       |  是       |
| 插入         |  O(n^2)        |  是     |  是        |
| 选择         |  O(n^2)        |  是     |  否        |


写好这三类排序，主要是要理清算法的过程。其中，选择排序涉及到交换，是一种不稳定的排序算法，

接下来的归并排序和快速排序用到了分治的思想，将问题分解为独立的子问题。

| 排序         |  时间复杂度    | 原地       | 稳定      |
| -----------  | -----------   | ---------- | -------  |
| 归并         |  O(nlog^n)     | 否       |  是        |
| 快排         |  O(nlog^n)     | 是       |  否        |

归并和快排虽然时间复杂度较低，但是各有优劣。归并不是原地排序算法，快排不稳定。要结合实际场景，灵活运用。

### 2.2.1 练习题

排序，主要要熟练归并和快排。其中，快排是每次只排一个元素的排序算法。练习题 [215. 数组中的第K个最大元素](https://leetcode.cn/problems/kth-largest-element-in-an-array/description/) 非常巧妙的结合数组下标通过快排只用 O(n) 的复杂度就找到了第 K 大元素。要注意的是，快排可以降序，也可以升序排列。这里找第 K 个最大元素，我们通过降序找下标为 k-1 的元素即可。同理，如果是第 K 个最小的元素，可以通过升序找 k-1 的元素。


当然，这题也可用堆排序，维护一个 K 个元素的大顶堆，堆顶即为第 K 个元素。堆在实时处理上比较好用，复杂度为 O(nlog^n)，对于 [215. 数组中的第K个最大元素](https://leetcode.cn/problems/kth-largest-element-in-an-array/description/) 还是用快排最合适。

## 2.2 查找

查找有几类，顺序查找，二分查找，散列表查找，二叉树查找等。

顺序查找的复杂度为 O(n)，二分查找为 O(log^n)，散列表查找为 O(1)，不过散列表需要额外的 O(n) 空间用来存储数据，二叉树查找，要构造二叉树在用二分查找来找，相对较为复杂。

二分查找往往用来有序数组的查找，通过数组查找。相比于二叉搜索树的查找对空间要求更高，虽然空间要求连续，但是存储数据量有限。相比于数组的二分查找，二叉搜索树的二分查找在空间上更有优势。

写好一个不出错的二分查找还是没那么容易的，要抓住两点，第一，左右位置要变，第二，跳出循环条件，包括等于，否则很有可能会陷入死循环。

### 2.2.1 练习题

二分的练习还是挺别出心裁，有意思的。

- [69. x 的平方根 ](https://leetcode.cn/problems/sqrtx/description/)，这是二分查找在数字中的体现，跳出常用的在数组上运用二分查找的逻辑。更有意思的点，它隐藏了二分查找的右区间。
- [153. 寻找旋转排序数组中的最小值](https://leetcode.cn/problems/find-minimum-in-rotated-sorted-array/description/)，这个还是运用二分查找，区别在于如何界定该往哪个区间查找。类似的题有 [0~n-1 中缺失的数字](./查找/arr/arr.go) 和 [数组中数值和下标相等的元素](./查找/arr/arr.go) 根据数组下标确定二分查找的区间。
- [实现模糊二分查找算法（比如大于等于给定值的第一个元素)] 和 [数字在排序数组中出现的次数](./查找/find/find.go) 这一题我们要注意，通过二分查找可以找有序数组中第一个和最后一个重复的元素。因此可以确定元素个数，这是二分查找的加强版。

当我们看到有序，本能的可以想想能否用二分查找实现。毕竟，二分快啊。

## 2.3 四种算法思想

# 3. 发散题

1. [238. 除自身以外数组的乘积](https://leetcode.cn/problems/product-of-array-except-self/)

这题有点像是动态规划，不过想是挺难想到的。可以列出几种 case，归纳统计，但是真正想到还是不容易的。


