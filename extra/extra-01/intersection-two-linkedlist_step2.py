class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        # リストの長さを計算する関数
        def getLength(head):
            length = 0
            current = head
            while current:
                length += 1
                current = current.next
            return length
        
        # 各リストの長さを取得
        lenA = getLength(headA)
        lenB = getLength(headB)

        # リストが交差→交差後は一致しなきゃいけないから長さに差がないはず
        currentA = headA
        currentB = headB
        # 長いリストの先頭を長さの差分だけ進める
        if lenA > lenB:
            for _ in range(lenA - lenB):
                currentA = currentA.next
        else:
            for _ in range(lenB - lenA):
                currentB = currentB.next
        
        # 2つのリストを同時に進めて、交差点を探す
        while currentA != currentB:
            currentA = currentA.next
            currentB = currentB.next
        
        return currentA  # 交差点、または None
