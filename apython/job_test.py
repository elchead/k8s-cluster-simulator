from parsing import *


def test_get_job_execution():
    job = Job()
    job.add_pod("owo", "zone2", PodData.withdata([10, 20, 30], [1, 2, 3]))
    job.add_pod("mowo", "zone2", PodData.withdata([10, 20, 30], [4, 5, 6]))
    assert job.get_execution_time() == 60


class TestGetPodRuns:
    job = Job()
    job.add_pod("mowo", "zone3", PodData.withdata([10, 20, 30], [4, 5, 6]))
    job.add_pod("owo", "zone2", PodData.withdata([10, 20, 30], [1, 2, 3]))
    it = tuple(job.get_pod_runs_for_plot())

    def test_first_item(self):
        zone, poddata = self.it[0]
        assert "zone2" == zone
        assert [1, 2, 3] == poddata.memory
        assert [10, 20, 30] == poddata.time

    def test_second_item(self):
        zone, poddata = self.it[1]
        assert "zone3" == zone
        assert [4, 5, 6] == poddata.memory
        assert np.array_equal([40, 50, 60], poddata.time) == True

    def test_time_is_not_mutated(self):
        assert self.job.node_data[1].time == [10, 20, 30]

    def test_nbr_migrations(self):
        assert self.job.nbr_migrations == 1

