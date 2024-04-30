class Solution:
    def findDisappearedNumbers(self, nums: List[int]) -> List[int]:
        #出現しなかった数記録するための配列
        disappeared_numbers = []
        #1~nそれぞれにfor文
        for i in range(1, len(nums) + 1):
            for num in nums:
                if num == i:
                    break
            else:
                disappeared_numbers.append(i)
        return disappeared_numbers
