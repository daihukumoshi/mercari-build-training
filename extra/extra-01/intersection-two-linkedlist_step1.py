class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        # ノードのアドレスを記録するためのセット
        nodes_in_A = set()
        
        # リストAを走査して、すべてのノードをセットに追加
        # currentをAの先頭に
        current = headA
        # currentがNoneになる(Aが終わる)までループ
        while current:
            #セット
            nodes_in_A.add(current)
            current = current.next
        
        # リストBを走査して、最初の共通ノードを探す（同様）
        current = headB
        while current:
            # Bの各nodeに対して、Aに含まれているnodeと一致しないか判定
            if current in nodes_in_A:
                return current
            current = current.next
        
        # 交差するノードがない場合はNoneを返す
        return None
