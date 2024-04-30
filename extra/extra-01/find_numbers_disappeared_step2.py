class Solution:
    def findDisappearedNumbers(self, nums: List[int]) -> List[int]: 
        #出現したかどうか記録するための配列を作り、初期値False
        is_present = [False] * len(nums)
        
        # nums に含まれる各要素に対して、対応するインデックスを True に設定
        for num in nums:
            is_present[num - 1] = True
        
        # False のままのインデックスを探し、対応する整数を結果のリストに追加
        disappeared_numbers = []
        for i in range(len(nums)):
            if not is_present[i]:
                disappeared_numbers.append(i + 1)
        
        return disappeared_numbers